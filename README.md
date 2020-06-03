# Hashpow

Fuck the hash proof of work for ctfer

## Online

https://hashpow.now.sh/
https://hashpow.now.sh/?c=66666&h=md5

## Usage

```
./hashpow --help

Usage of ./build/hashpow:
  -c string (**require**)
        code
  -t string (**require**)
        hash type : md5 sha1
  -p int
        starting position of hash
  -pf string
        text prefix
  -sf string
        text suffix

  -port int
        Web server port (default 3000)
  -s    Run as a web server provide api
```

## Cli

`./hashpow -c code -t [md5,sha1] [-p pos -pf prefix -sf suffix]`

## Server

`./hashpow -s -port 3000`

It set timeout 10s. If you get **Empty reply from server**, that will be timeout!

### API

Request

```
/hashpow?c=code
/hashpow?c=code&p=pos&t=hash&pf=prefix&sf=suffix
/hashpow?c=aaaaaa&t=md5
```

Response

```
{
      "code":0,
      "msg":"success",
      "data":{
            "code":"aaaaaa",
            "hash":"md5",
            "msg":"success",
            "pos":0,
            "prefix":"",
            "result":"dsEfACYS",
            "suffix":""
      }
}
```