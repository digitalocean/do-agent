qdisc [![Build Status](https://github.com/ema/qdisc/actions/workflows/makefile.yml/badge.svg)](https://github.com/ema/qdisc/actions/)
=====

Package `qdisc` allows getting queuing discipline information via netlink,
similarly to what `tc -s qdisc show` does.

Example usage
-------------

    package main

    import (
        "fmt"

        "github.com/ema/qdisc"
    )

    func main() {
        info, err := qdisc.Get()

        if err == nil {
            for _, msg := range info {
                fmt.Printf("%+v\n", msg)
            }
        }
    }
