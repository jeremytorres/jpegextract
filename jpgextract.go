/*
 Copyright (c) 2013 Jeremy Torres, https://github.com/jeremytorres/jpegextract

 Permission is hereby granted, free of charge, to any person obtaining
 a copy of this software and associated documentation files (the
 "Software"), to deal in the Software without restriction, including
 without limitation the rights to use, copy, modify, merge, publish,
 distribute, sublicense, and/or sell copies of the Software, and to
 permit persons to whom the Software is furnished to do so, subject to
 the following conditions:

 The above copyright notice and this permission notice shall be
 included in all copies or substantial portions of the Software.

 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND,
 EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF
 MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND
 NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE
 LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION
 OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION
 WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
*/

// Command utilizing the RawFileparser library for parsing Nikon Electronic Files (RawFiles).
// The embedded JPEGs are extracted and optionally:
// (1) rotated via oritenation provided in RawFile EXIF info;
//
// Usage:
//     jpgextract --raws "NEF,CR2" --dest-dir "/path_to/output_dir"
//                --src-dirs "/path_to/source1,/path_to/source2"
//                [--num-routines "8" --quality "80" --rotate]
package main

import (
	"github.com/codegangsta/cli"
	"github.com/jeremytorres/rawparser"
	"log"
	"math"
	"os"
	"os/exec"
	"runtime"
	"runtime/pprof"
	"strconv"
	"strings"
	"time"
)

// RawFileParserPair is a struct for containing a list of files
// for a specific parser.
type RawFileParserPair struct {
	file   string
	parser rawparser.RawParser
}

const (
	// RawTypesKey is the constant representing the command line argument for the RAW file
	// types to be processed.
	RawTypesKey = "raws"

	// SrcDirKey is the constant representing the command line argument for the source directories
	// containing the source RawFiles.
	SrcDirKey = "src-dirs"
	// DestDirKey is a constant representing the command line argument for the destination directory
	// where extracted jpegs shall be placed.
	DestDirKey = "dest-dir"
	// NumRoutinesKey is the constant representing the command line argument for the max number of
	// go routines to be processed.  As this utility is mostly file IO bound, it is important
	// not to use too many routines ("too many files open..." errors can occur from OS).  Therefore,
	// if this value is greater than the logical cores, a warning is logged and the value is set to
	// the number of logical cores.
	NumRoutinesKey = "num-routines"
	// QualityKey is the constant representing the command line argument for the JPEG quality used
	// when processing the embedded JPEG in a RawFile.
	QualityKey = "quality"
	// RotateKey is the constant representing the command line argument indicating rotation of
	// JPEGs should occur based on the EXIF info embedded within the RawFile.  ImageMagic's 'convert'
	// utility is used and is checked at startup for existence.
	RotateKey = "rotate"
	// AppVersionKey is the constant defining the current version of this command utility.
	AppVersionKey = "1.0"
	// ImageMagicConvertBin is the constant representing ImageMagic's 'convert' utility.
	ImageMagicConvertBin = "convert"
)

var (
	destDir, sqlLiteDb     string
	rawFileExts, srcDirs   []string
	numOfRoutines, quality int
	rotate                 bool
	parsers                *rawparser.RawParsers
	// validParserKeys is a slice of RAW file parsers supported by this implementation.
	validParserKeys = []string{
		rawparser.NefParserKey,
		rawparser.Cr2ParserKey,
	}
)

// processFilesConcurrent processes RawFiles concurrently, based on the number goroutines
// specified.
func processFilesConcurrent(rp *RawFileParserPair, c chan<- bool) {

	go func(rp *RawFileParserPair, c chan<- bool) {
		rawfile, err := rp.parser.ProcessFile(&rawparser.RawFileInfo{rp.file, destDir, quality, numOfRoutines})
		if err != nil {
			log.Printf("Error processing file: '%s' error: %v\n", rp.file, err)
		} else {
			if rotate && rawfile.JpegOrientation != 0.0 {
				// rotate jpeg
				go func(fileName string, radiansCw float64) {
					degrees := radiansCw * (180 / math.Pi)
					log.Printf("Rotating image %f degrees for jpeg: '%s'\n", degrees, fileName)
					cmd := exec.Command(ImageMagicConvertBin, "-rotate", strconv.FormatFloat(degrees, 'f', 2, 64), fileName, fileName)
					err := cmd.Start()
					if err != nil {
						log.Fatal(err)
					}
					err = cmd.Wait()
					if err != nil {
						log.Printf("Command finished with error: %v", err)
					}
				}(rawfile.JpegPath, rawfile.JpegOrientation)
			}
		}
		// signal completion of work
		c <- true
	}(rp, c)
}

func isRawFileExtValid(ext string) bool {
	for _, validExt := range validParserKeys {
		if ext == validExt {
			return true
		}
	}
	return false
}

// processCli parses command line arguments and checks for validity of user-specified
// values.
func processCli() bool {
	var processed = false

	app := cli.NewApp()
	app.Name = "jpgextract"
	app.Usage = "Processes RAW files and extracts JPEGs and optionally create a SQLite database."
	app.Version = AppVersionKey
	app.Author = "Jeremy Torres"
	app.Flags = []cli.Flag{
		cli.StringFlag{RawTypesKey, "", "comma-separated list of RAW file extensions to process.  Current supported " +
			"RAW files: [NEF | CR2]"},
		cli.StringFlag{SrcDirKey, "", "comma-separated list of full paths to the directories containing RawFiles " +
			"(e.g., /path_to/dir1,/path_to/dir2..."},
		cli.StringFlag{DestDirKey, "", "the full path to the directory containing extracted jpegs"},
		cli.IntFlag{NumRoutinesKey, 2, "the number of concurrent files to be processed"},
		cli.IntFlag{QualityKey, 80, "JPEG encoding quality used for extracted jpegs"},
		cli.BoolFlag{RotateKey, "ImageMagic's 'convert' command will be used to rotate jpegs based on EXIF info from RawFile.  'convert' must be in PATH"},
	}
	app.Action = func(c *cli.Context) {
		rawExts := strings.TrimSpace(c.String(RawTypesKey))
		srcDir := strings.TrimSpace(c.String(SrcDirKey))
		destDir = strings.TrimSpace(c.String(DestDirKey))
		numOfRoutines = c.Int(NumRoutinesKey)
		quality = c.Int(QualityKey)
		rotate = c.Bool(RotateKey)

		// src and dest dirs required; remaing args have sane defaults
		if rawExts == "" || srcDir == "" || destDir == "" {
			cli.ShowAppHelp(c)
			os.Exit(1)
		}

		if rotate && !isImagicConvertInPath(ImageMagicConvertBin) {
			log.Fatal("Rotation of jpegs was enables, but ImageMagic's 'convert' is not in path!")
			exitWithErr()
		}

		// verify RAW file extensions
		rawFileExts = strings.Split(rawExts, ",")
		for i, ext := range rawFileExts {
			rawFileExts[i] = strings.ToUpper(strings.TrimSpace(ext))
			if !isRawFileExtValid(rawFileExts[i]) {
				log.Printf("Error: Invalid Raw File extension: %s\n", rawFileExts[i])
				exitWithErr()
			}
		}

		// verify user-provided src dirs
		srcDirs = strings.Split(srcDir, ",")
		//log.Printf("SrcDirs: %s", srcDirs)
		for i, dir := range srcDirs {
			srcDirs[i] = strings.TrimSpace(dir)
			// ensure directory ends with '/'
			if !strings.HasSuffix(dir, "/") {
				srcDirs[i] += "/"
			}
			isValid, errStr := validateUserDir(srcDirs[i])
			if !isValid {
				log.Println("Error: ", errStr)
				exitWithErr()
			}
		}

		// verify user-provided dest dir
		isValid, errStr := validateUserDir(destDir)
		if !isValid {
			log.Println(errStr)
			exitWithErr()
		}
		// ensure directory ends with '/'
		if !strings.HasSuffix(destDir, "/") {
			destDir += "/"
		}

		processed = true
	}

	app.Run(os.Args)

	return processed
}

func initParsers() {
	parsers = rawparser.NewRawParsers()
	cr2Parser, cr2Key := rawparser.NewCr2Parser(isHostLittleEndian())
	parsers.Register(cr2Key, cr2Parser)

	nefParser, nefKey := rawparser.NewNefParser(isHostLittleEndian())
	parsers.Register(nefKey, nefParser)
}

func doProcess() int {
	log.Printf("RawTypes: %v SourceDirs: %v DestinationDir: %s JPEG Quality: %d Rotate Images: %v\n",
		rawFileExts, srcDirs, destDir, quality, rotate)

	total := 0
	done := make(chan bool)

	// process all src dirs
	for _, dir := range srcDirs {
		// process all raw file types
		for i, rawType := range rawFileExts {
			globPattern := dir + "*." + rawType
			files, _ := getFilesForExt(globPattern)
			fileCnt := len(files)

			if fileCnt > 0 {
				log.Printf("Raw Type: %s ==> Processing '%s' %d files with %d NumCPU:\n",
					rawFileExts[i], dir, len(files), runtime.NumCPU())

				drainCnt := 1
				routinesActive := 0

				for j := 0; j < fileCnt; j++ {
					// keep the specified number of go routines active
					if routinesActive == numOfRoutines {
						// Completion of routines occur in any order.  Count the completion signals
						// by draining channel after launching routines.
						// Drain the channel.
						for c := 0; c < drainCnt; c++ {
							<-done // wait for task to complete
							routinesActive--
						}
					}

					processFilesConcurrent(&RawFileParserPair{files[j], parsers.GetParser(rawType)}, done)
					routinesActive++
				}
				total += fileCnt
			}
		}
	}

	close(done)

	return total
}

func setup() {
	success := processCli()
	if success {
		/*
			// don't allow num of channels > logical cores
			if numOfRoutines > runtime.NumCPU() {
				log.Printf("Note: Processing %d concurrently, as %d exceeds CPU count of host.\n",
					runtime.NumCPU(), numOfRoutines)
				numOfRoutines = runtime.NumCPU()
			}

		*/
		runtime.GOMAXPROCS(numOfRoutines)

		initParsers()
	} else {
		exitWithErr()
	}
}

func main() {
	f, err := os.Create("/tmp/jpgextract_cpu.prof")
	if err != nil {
		log.Fatal(err)
	}
	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	t0 := time.Now()

	setup()

	cnt := doProcess()

	duration := time.Since(t0)

	log.Printf("jpgextract processed %d files in %02f mintues (%02f seconds)\n",
		cnt, duration.Minutes(), duration.Seconds())
}