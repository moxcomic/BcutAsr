# BcutAsr
Bcut Audio to Text Api for Golang

# Usage
`import "github.com/moxcomic/bcutasr"`

```golang
res, err := New().Parse("./1.mp3")
if err != nil {
    panic(err)
}
```