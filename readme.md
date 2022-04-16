### aowoo (cat's meow sound but in baby tiger way.  `嗷呜～` in Chinese)

sound paly in multiple platform. 

1. linux 
2. mac os 
3. ios 
4. android 
5. javascript
6. windows


simple example

    func main() {

        path := filepath.Join("assets", "2.wav")
        f, err := os.Open(path)
        if err != nil {
            log.Fatal(err)
        }
        defer f.Close()

        aowoo.Open(44100, 16, 2)
        aowoo.NewSource(f, 1, true)

        ch := make(chan int, 1)
        <-ch
    }
