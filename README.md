# Compare Cron Jobs between two configurations

This package provides a simple way to compare two cron job configurations. Solve issue when you have to maintain
multiple cron jobs configurations and need to know what is different between them.

Comparing is based on command, name and schedule and result is presented in a table.

## Installation

Basic use

```bash
go run github.com/jakubfrieb/go-utils-compare@latest \ 
    ../magento-values.yaml \
    ../magento-values.yaml
```
Automated processing
For json output you can add --json flag

```bash 
go run github.com/jakubfrieb/go-utils-compare@latest \
    --json \
    ../magento-values.yaml \
    ../magento-values.yaml
```


