## Nodectl

This is an open source Go rewrite of the `nodectl` command used for out of band management of Mixtile's [Cluster Box](https://www.mixtile.com/cluster-box/).

Advantages over the version that ships with the hardware:
* There is additional error checking with extensive logging of where problems are
* The `poweroff` command is fully supported
* The module can be used in other applications directly rather than calling the CLI
