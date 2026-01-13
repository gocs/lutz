# look up time zone

```
...
<continent>/<city>/<subcategory> <hour> <minutes>
<continent>/<city>/<subcategory> <hour> <minutes>
<continent>/<city>/<subcategory> <hour> <minutes>
<continent>/<city>/<subcategory> <hour> <minutes>
...
```

dl timezone from iana (i.e. eggert/tz)
parse for each continent
merge data in one sorted file

- download through http request
- unzip a compressed gzip file in memory
- iterate through all files (only continents)
    - every iteration instatiates a scanner
    - every line is parsed to get a specific format
    - every line is sorted per insert
    - every after parsing is appended to the lookup file

## building

```
go build ./...
```

## cron

the `tz` file is gonna be updated every month