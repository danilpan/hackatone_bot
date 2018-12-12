package post

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"github.com/finalist736/finalistx-tg-bot"
)

const postURL = "http://finalistx.com/email.php"

func SendPost(data *finbot.CourseSign) error {

	params := fmt.Sprintf(
		"name=%s&email=%s&tel=%s&course=%s",
		data.Name,
		data.Email,
		data.Telephone,
		data.Course)

	buf := bytes.NewBufferString(params)
	resp, err := http.Post(
		postURL,
		"application/x-www-form-urlencoded",
		buf)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	ba, err := ioutil.ReadAll(resp.Body)
	fmt.Printf("response: %s\n", ba)

	if resp.StatusCode != 200 {
		return errors.New("not 200 response")
	}
	return nil
}
