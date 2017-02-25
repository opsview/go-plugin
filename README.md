# go-plugin

[![GoDoc](https://godoc.org/github.com/ajgb/go-plugin?status.png)][godoc]
[![Build Status](https://travis-ci.org/ajgb/go-plugin.svg?branch=master)][travis]
[![GPLv3](https://img.shields.io/badge/licence-GPLv3-green.svg)][license]

[travis]: https://travis-ci.org/ajgb/go-plugin
[license]: https://github.com/ajgb/go-plugin/blob/master/LICENSE
[godoc]: https://godoc.org/github.com/ajgb/go-plugin

## Description

go-plugin is a library for creating monitoring plugins

## Synopsis

    package main

    import (
      "github.com/ajgb/go-plugin"
    )
    
    var opts struct {
      Hostname string `short:"H" long:"hostname" description:"Host" default:"localhost"`
      Port     int    `short:"p" long:"port" description:"Port" default:"123"`
    }

    func main() {
      check := plugin.New("check_service", "1.0.0")
      defer check.Final()

      check.AddMessage("Service %s:%d", opts.Hostname, opts.Port)

      serviceMetrics, err := .... // retrieve metrics
      if err != nil {
        check.ExitCritical("Connection to %s:%d failed: %s", opts.Hostname, opts.Port, err)
      }
    
      for metric, value := range serviceMetrics {
        check.AddMetric(metric, value)
      }
    }
