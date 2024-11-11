# cnamex

**cnamex** is a tool to check if subdomains have CNAME records or not.

## Installation

To install **cnamex**, run the following command:

```bash
go install github.com/hackruler/cnamex@latest
```
## Usage 

You can pipe the input and it will store the subdomains to different files which has any CNAME and not any CNAME.

```bash
cat subdomains.txt | cnamex 
```

```bash
cnamex <input file>
```

It will store the subdomains which has anycname to `cnames_found.txt`

And

It will store the subdomains which has not any cname to `no_cname-found.txt`

