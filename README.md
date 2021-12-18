# The URL size fetcher

## What is this?
This is an example application to show how concurrency works.  
The application receives urls as arguments and runs workers to fetch its pages sizes.  
Response wil be sorted by size descending

Yeah, I know that go 1.14 too old in 2k21, but im too lazy to update in on my personal laptop :)

## Example

```text
./main https://google.com https://yahoo.com https://bing.com https://ya.ru https://alwaysunknownwebsite
+---+------------------------------+----------------------------------------------------------------------------------------------------------------------------------------+
| # | URL                          | RESULT                                                                                                                                 |
+---+------------------------------+----------------------------------------------------------------------------------------------------------------------------------------+
| 0 | https://yahoo.com            | 136 kB                                                                                                                                 |
| 1 | https://bing.com             | 32 kB                                                                                                                                  |
| 2 | https://ya.ru                | 17 kB                                                                                                                                  |
| 3 | https://google.com           | 6.9 kB                                                                                                                                 |
| 4 | https://alwaysunknownwebsite | execution error: error executing http request: Get "https://alwaysunknownwebsite": dial tcp: lookup alwaysunknownwebsite: no such host |
+---+------------------------------+----------------------------------------------------------------------------------------------------------------------------------------+%  
```


