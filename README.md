Here's the corrected version:

# Tools for Dealing with ASCIINEMA Casts

There is an excellent tool made by [Ciro S. Costa](https://github.com/cirocosta/asciinema-edit). Unfortunately, it has not been maintained for a while, so I had to update it to be usable.
To simplify things, I added `record` and `play` commands to the tool.

- [Tools for Dealing with ASCIINEMA Casts](#tools-for-dealing-with-asciinema-casts)
  - [Usage](#usage)
    - [Quantize](#quantize)
    - [Speed](#speed)
    - [Cut](#cut)
    - [Record](#record)
    - [Play](#play)

`asciinema-edit` is a tool whose purpose is to post-process ASCIINEMA casts (V2), either from [asciinema](https://github.com/asciinema/asciinema) itself or [termtosvg](https://github.com/nbedos/termtosvg).

Three transformations have been implemented so far:

- [`quantize`](#quantize): Updates the cast delays following quantization ranges.
- [`cut`](#cut): Removes a certain range of time frames.
- [`speed`](#speed): Updates the cast speed by a certain factor.
- [`record`](#record): Records the cast.
- [`play`](#play): Plays the cast.

With these tools, you can improve your cast by:

- Speeding up parts that are not very important.
- Reducing delays between commands.
- Completely removing parts that don't add value to the cast.

## Usage

### Quantize

``` sh
NAME:
   asciinema-edit quantize - Updates the cast delays following quantization ranges.

   The command acts on the delays between the frames, reducing such
   timings to the lowest value defined in a given range that they
   lie in.

   For instance, consider the following timestamps:

      1  2  5  9 10 11

   Assuming that we quantize over [2,6], we would cut any delays between 2 and
   6 seconds to 2 second:

      1  2  4  6  7  8

   This can be more easily visualized by looking at the delay quantization:

      delta = 1.000000 | qdelta = 1.000000
      delta = 3.000000 | qdelta = 2.000000
      delta = 4.000000 | qdelta = 2.000000
      delta = 1.000000 | qdelta = 1.000000
      delta = 1.000000 | qdelta = 1.000000

   If no file name is specified as a positional argument, a cast is
   expected to be served via stdin.

   Once the transformation has been performed, the resulting cast is
   either written to a file specified in the '--out' flag or to stdout
   (default).

EXAMPLES:
   Make the whole cast have a maximum delay of 2s:

     asciinema-edit quantize --range 2 ./123.cast

   Make the whole cast have time delays between 300ms and 1s cut to
   300ms, delays between 1s and 2s cut to 1s and any delays bigger
   than 2s, cut down to 2s:

     asciinema-edit quantize --range 0.3,1 --range 1,2  --range 2 ./123.cast

USAGE:
   asciinema-edit quantize [ command options] [ filename ]

OPTIONS:
   --range value  quantization ranges ( comma delimited )
   --out value    file to write the modified contents to
   
```

### Speed

```sh
NAME:
   asciinema-edit speed - Updates the cast speed by a certain factor.

   If no file name is specified as a positional argument, a cast is
   expected to be served via stdin.

   If no range is specified ( start=0, end=0 ), the whole event stream
   is processed.

   Once the transformation has been performed, the resulting cast is
   either written to a file specified in the '--out' flag or to stdout
   (default).

EXAMPLES:
   Make the whole cast ( "123.cast" ) twice as slow:

     asciinema-edit speed --factor 2 ./123.cast

   Cut the duration in half:

     asciinema-edit speed --factor 0.5 ./123.cast

   Make only a certain part of the video twice as slow:

     asciinema-edit speed --factor 2  --start 12.231 --factor 45.333  ./123.cast

USAGE:
   asciinema-edit speed [command options] [filename]

OPTIONS:
   --factor value  number by which delays are multiplied by (default: 0)
   --start value   initial frame timestamp (default: 0)
   --end value     final frame timestamp (default: 0)
   --out value     file to write the modified contents to
```


### Cut

```sh
NAME:
   asciinema-edit cut - Removes a certain range of time frames.

   If no file name is specified as a positional argument, a cast is
   expected to be served via stdin.

   Once the transformation has been performed, the resulting cast is
   either written to a file specified in the '--out' flag or to stdout
   (default).

EXAMPLES:
   Remove frames from 12.2s to 15.3s from the cast passed in the commands
   stdin.

     cat 1234.cast | asciinema-edit cut --start=12.2 --end=15.3

   Remove the exact frame at timestamp 12.2 from the cast file named
   1234.cast.

     asciinema-edit cut  --start=12.2 --end=12.2   1234.cast

USAGE:
   asciinema-edit cut [command options] [filename]

OPTIONS:
   --start value  initial frame timestamp (required) (default: 0)
   --end value    final frame timestamp (required) (default: 0)
   --out value    file to write the modified contents to
```

### Record

``` sh
NAME:
   asciinema-edit record -  Records cast to an output file .

EXAMPLES:

   asciinema-edit rec  ./123.cast

USAGE:
   asciinema-edit record [command options] [filename]

OPTIONS:
   --args value   shell command arguments
   --shell value  shell command  [$SHELL]

```

### Play

``` sh 
NAME:
   asciinema-edit play - Plays a cast from a file .

EXAMPLES:

   asciinema-edit play  ./123.cast

USAGE:
   asciinema-edit play [command options] [filename]

OPTIONS:
   --speed value  speed of playback (default: 1)
```
