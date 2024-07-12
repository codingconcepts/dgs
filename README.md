# dgs
A streaming version of dg, which writes data directly to a database without any kind of buffering.

### Installation

Find the release that matches your architecture on the [releases](https://github.com/codingconcepts/dgs/releases) page.

Download the tar, extract the executable, and move it into your PATH. For example:

```sh
tar -xvf dgs_0.0.1_macos_amd64.tar.gz
```

### Usage

dgs uses cobra for managing commands, of which there are currently 2:

```
Usage:
  dgs gen [command]

Available Commands:
  config      Generate the config file for a given database schema
  data        Generate relational data
```

### Generate config

If familiar with dgs configuration, you may prefer to hand-roll your dgs configs. However, if you'd prefer to use dgs itself to generate the configuration for you, you can use `dgs gen config` to generate a configuration file for you.

Note that this tool will sort the tables in the config file in dependency order using Kahn's algorithm to determin topological order (guaranteeing that tables with a reference to another table will be generated after the table they depend reference).

```sh
dgs gen config \
--url "postgres://root@localhost:26257?sslmode=disable" \
--schema public > examples/e-commerce/config.yaml
```

### Generate data

Once you have a dgs config file, you can generate data

```sh
dgs gen data \
--config examples/e-commerce/config.yaml \
--url "postgres://root@localhost:26257?sslmode=disable" \
--workers 4 \
--batch 10000
```

### Todo

- [ ] [Performance] Consider sorting data by primary key column(s) before inserting
