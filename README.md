# go-google-search

Google search crawler with NSQ worker.

## Usage

Example usage from command line:

```sh
# Install command line application for crawl scheduling
$ go install github.com/crackcomm/crawl/nsq/crawl-schedule
# Schedule crawl of google search results
$ crawl-schedule \
      --topic google_search \
      --callback github.com/crackcomm/go-google-search/spider.Google \
      "https://www.google.com/search?q=Github"
```

## License

                                 Apache License
                           Version 2.0, January 2004
                        http://www.apache.org/licenses/

## Authors

* [Łukasz Kurowski](https://github.com/crackcomm)
