# jpegextract [![Build Status](https://travis-ci.org/jeremytorres/jpegextract.png)](https://travis-ci.org/jeremytorres/jpegextract) [![GoDoc](https://godoc.org/github.com/jeremytorres/jpegextract?status.png)](http://godoc.org/github.com/jeremytorres/jpegextract)

## Overview
jpegextract is a command-line utility to extract the embedded JPEGs from a camera RAW file.  It's a GO language command based on [RawParser](https://github.com/jeremytorres/rawparser).  There are existing tools that perform this or similar functionality; however, the reasons for creating this tool:

1. I have many RAW files that are processed using commercial software, yet on occassion, I would like the camera-produced JPEG for comparison.
2. To utilize the concurrency model provided by the [GO](http://golang.org) language to process multiple files without any explicit "traditional" locking (e.g, mutexes)
3. Experiment with GO's "C" package and interfacing with existing C libraries.

## Dependencies
* GO 1.2 (_maybe_ older GO 1.1.2? but not tested)
* [libjpeg](http://www.ijg.org)
* [ImageMagick](http://www.imagemagick.org/)
    * `convert` utility Required for rotating JPEGs
* Optional (highly-recommended):
    * [TurboJpeg](http://www.libjpeg-turbo.org/)
    * If you have many JPEGs to extract, TurboJpeg provides noticebly better performance.  See performance [observations](#performance).
 
### Why not only use GO's [image/jpeg](http://golang.org/pkg/image/jpeg/) package?
I originally planned on making the utility pure GO. In fact for small batches of images, image/jpeg does the job.  However, when you have a _lot_ of files to process, C-based image libraries are hard to beat with respect to processing time--even more, using TurboJpeg with its optimized processor instructions.  See [performance](#performance-tests) to further understand why.

## Usage
* Obtain the utility.  The following will retrieve _jpegextract_ and any dependencies and will build the _default_ executable.  The _default_ uses GO's image/jpeg package.
 
```go
go get github.com/jeremytorres/jpegextract
```

   * jpegextract is command-line executable: $GOPATH/bin/jpegextract

* To utilize libjpeg or turbojpeg:

```bash
cd $GOPATH/src/github.com/jeremytorres/jpegextract

# Build using libjpeg
# See jpeglibjpeg.go for CFLAGS and LDFLAGS and update for your environment.  # The defaults for Linux and Mac OS (10.9) should work.
go build -tags jpeg

# Build using turbojpeg
# See jpegturbojpeg.go for CFLAGS and LDFLAGS and update for your environment.  # The defaults for Linux and Mac OS (10.9) should work.
go build -tags turbojpeg
```
   
* Invoke `jpegextract`
    * Get help.  Displays information on parameters and defaults.
```bash
./jpegextract -h
```

    * Execute utility.

```bash
./jpgextract --raws "CR2,NEF" --src-dirs "/path_to/raw_files/dir1,/path_to/raw_file/dir" --dest-dir "/path_to/extract_dir" --num-routines 8 --quality 75
```

   * Verify output.  Extracted JPEGs will contain the source RAW file name with an "_extracted.jpg" extension.

### `--quality` parameter
The quality parameter is the JPEG compressors "quaility" value.  This is a value from 0-100, where 0 is the lowest compression quality.  This value has a dramatic impact on overall processing time and the size of the resulting extracted JPEG.  A good rule of thumb: 80 or less for really good JPEGs.  If storage size is of concern, us a smaller value.

### `--num-routines` parameter
This defines the max number of files to process concurrently (setting GOMAXPROCS internally).  In the above example, 8 files would be processed simultaneously.  It's recommended to set this parameter to the maximum number of logical cores of the host machine; however, your mileage _will_ vary!  Disk IO performance will drive the perforance gains of this parameter, so experiment with this parameter to tweak for your environemt.

### Current RAW file format support
* Nikon NEF
    * Nikon D700
* Canon CR2
    * Canon 5D MarkII
    * Canon 1D MarkIII

### Performance
As is always the case, performance *will* very.  Below are _my_ observations based upon the aforementioned [dependencies](#dependencies).

#### Performance Tests
##### Test Setup
###### Test Machine
- Mac Pro (Early 2008)
    - OS: Mac OS X 10.9.1 "Mavericks"
    - Processor: 2 x 2.8GHz Quad-Core Intel Xeon
    - Memory: 16 GB 800 MHz DDR2 memory
    - Disk (OS): 128 GB SSD
    - Disk (Data): 3 TB software RAID comprised of 3 1TB 7200RPM hard drives

##### Test 1: TurboJpeg
- TurboJpeg library installed via [homebrew](http://brew.sh/)
- _CFLAGS_ for wrapper code: -`O2`
- _jpegextract_ Utility Config
    - TurboJpeg defined in jpeg.go
    - `LDFLAGS=-L/path_to/jpeg-turbo/lib -lturbojpeg`
    - `CFLGAS=-I/path_to/jpeg-turbo/include`
_ Inputs to _jpegextract_ Utility
    - 1745 Canon CR2 files: a mixture of 10 and 21 mega pixel files.
    - Default JPEG compression value of 80.
    - [GOMAXPROCS](http://golang.org/pkg/runtime/#GOMAXPROCS) specified: 8
- Results
    - Time to Complete: *58.67* seconds (average per file: *.03* seconds)

##### Test 2: libjpeg
- libjpeg used is OS X default for 10.9
    - /usr/lib/libjpeg
- _CFLAGS_ for wrapper code: -`O2`
- _jpegextract_ Utility Config
    - libjpeg defined in jpeg.go
    - `LDFLAGS=-ljpeg`
_ Inputs to _jpegextract_ Utility
    - 1745 Canon CR2 files: a mixture of 10 and 21 mega pixel files.
    - Default JPEG compression value of 80.
    - [GOMAXPROCS](http://golang.org/pkg/runtime/#GOMAXPROCS) specified: 8
- Results
    - Time to Complete: *136.28* seconds (average per file: *.08* seconds)

##### Test 3: GO image/jpeg
- GO version 1.2
- _jpegextract_ Utility Config
_ Inputs to _jpegextract_ Utility
    - 1745 Canon CR2 files: a mixture of 10 and 21 mega pixel files.
    - Default JPEG compression value of 80.
    - [GOMAXPROCS](http://golang.org/pkg/runtime/#GOMAXPROCS) specified: 8
- Results
    - Time to Complete: *470.29* seconds (average per file: *.27* seconds)

### Current Development Status
- I consider the current status a beta version as there is a laundry list of this I will like to support:
    - Add performance benchmarks
    - Add additional camera RAW file support
    - Create a "better" parser interface

