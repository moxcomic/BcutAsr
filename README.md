# BcutAsr
Bcut Audio to Text Api for Golang

# Usage
`import "github.com/moxcomic/bcutasr"`

```golang
res, err := bcutasr.New().Parse("./1.mp3")
if err != nil {
    panic(err)
}
```

The result is obtained by referring to `result.data`