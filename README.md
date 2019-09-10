# Hashpow

Fuck the hash proof of work for ctfer

## Usage

```
./hashpow --help

Usage of ./build/hashpow:
  -c string
        code
  -h string
        hash type : md5 sha1
  -p int
        starting position of hash
  -pf string
        text prefix
  -port int
        Web server port (default 3000)
  -s    Run as a web server provide api
  -sf string
        text suffix
```

## Cli

`./hashpow -c code [-p pos -h [md5,sha1] -pf prefix -sf suffix]`

## Server

`./hashpow -s -port 3000`

It set timeout 10s. If you get **Empty reply from server**, that will be timeout!

### API

Request

```
/hashpow?c=code
/hashpow?c=code&p=pos&h=hash&pf=prefix&sf=suffix
```

Response

```
{"code":"aaaaaa","hash":"","msg":"success","pos":0,"prefix":"","result":"dsEfACYS","suffix":""}
```