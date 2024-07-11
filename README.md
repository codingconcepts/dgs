# dgs
A streaming version of dg, which writes data directly to a database without any kind of buffering.

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/dgs/releases) page.

Download the tar, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf dgs_0.0.1_macos_amd64.tar.gz
```

### Usage

```sh
dgs --help
Usage dgs:
  -batch int
        query and insert batch size (default 10000)
  -config string
        absolute or relative path to the config file
  -debug
        enable debug logging
  -url string
        database connection string
  -workers int
        number of workers to run concurrently (default 4)
```

### Todo

- [ ] [Bug] Add length field to range (to prevent Int63n from failing because of max - min = 0 error)

- [ ] [Performance] Process ref dependency tables first and run them concurrently
- [ ] [Performance] Run inserts in parallel
- [ ] [Performance] Use ints for min and max ranges where possible
- [ ] [Performance] Consider sorting data by primary key column(s) before inserting
