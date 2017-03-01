package scanpay
import(
    "bytes"
    "crypto/hmac"
    "crypto/sha256"
    "crypto/subtle"
    "encoding/base64"
    "encoding/json"
    "errors"
    "io"
    "net/http"
)

type opts struct {
    CardHolderIP string
}

func (c *Client) req(uri string, in interface{}, out interface{}, opts *opts) error {
    var inRdr io.Reader
    reqtype := "GET"
    if in != nil {
        bdata, err := json.Marshal(in)
        if err != nil {
            return err
        }
        inRdr = bytes.NewReader(bdata)
        reqtype = "POST"
    }
    proto := "https://"
    if c.insecure {
        proto = "http://"
    }
    req, err := http.NewRequest(reqtype, proto + c.host + uri, inRdr)
    if err != nil {
        return err
    }
    req.Header.Set("Authorization", "Basic " + base64.StdEncoding.EncodeToString([]byte(c.apikey)) )
    if in != nil {
        req.Header.Set("Content-Type", "application/json; charset=UTF-8")
    }
    if opts != nil && opts.CardHolderIP != "" {
        req.Header.Set("X-Cardholder-Ip", opts.CardHolderIP)
    }
    res, err := c.Do(req)
    if err != nil {
        return err
    }
    defer res.Body.Close()
    if res.StatusCode != 200 {
        if res.StatusCode == 403 {
            return errors.New("Invalid scanpay client apikey")
        }
        return err
    }
    if err := json.NewDecoder(io.LimitReader(res.Body, 131072)).Decode(out); err != nil {
        return err
    }
    return nil
}

/* Debug-only method, may change */
func (c *Client) MakeInsecure() {
    c.insecure = true
}

/* Debug-only method, may change */
func (c *Client) SetHost(host string) {
    c.host = host
}

func (c *Client) signatureIsValid(req *http.Request, body []byte) bool {
    hmacsha2 := hmac.New(sha256.New, []byte(c.apikey))
    hmacsha2.Write(body)
    rawSig := hmacsha2.Sum(nil)
    buf := make([]byte, base64.StdEncoding.EncodedLen(len(rawSig)))
    base64.StdEncoding.Encode(buf, rawSig)
    return subtle.ConstantTimeCompare(buf, []byte(req.Header.Get("X-Signature"))) == 1
}
