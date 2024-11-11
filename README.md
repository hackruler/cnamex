# cnamex

**cnamex** is a tool to check if subdomains have CNAME records or not.

## Installation

To install **cnamex**, run the following command:

```bash
go install github.com/hackruler/cnamex@v1.0.6
```
## Usage 

You can pipe the input and it will store the subdomains to different files which has any CNAME and not any CNAME.

```bash
cat subdomains.txt | cnamex 
```

```bash
cnamex -f <input file>
```



