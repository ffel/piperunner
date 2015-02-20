Pipe Runner
===========

Run a pool of workers that run external commands.

These command accept input on `stdin` and produce output on `stdout`.

My main objective is to have a pool of workers that run a limited number
of [pandoc](http://johnmacfarlane.net/pandoc/) instances in parallel.

> pandoc is used in the tests.

Future wishes

-   better error messages in case of non existing commands

-   ability to increase the number of worker
