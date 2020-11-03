package main

import (
	"GameNerdzMonitor/http"
	"GameNerdzMonitor/models"
	"bufio"
	"encoding/csv"
	"errors"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	// TESTING
	DARKNESS_ABLAZE_BOOSTER_BOX        = "42004"
	VIVID_VOLTAGE_BOOSTER_BOX_PREORDER = "43220"

	// COPS
	HIDDEN_FATES_ETB                               = "43748"
	KANTO_POWER_COLLECTION                         = "43229"
	CHAMPIONS_PATH_ETB                             = "42635"
	CHAMPIONS_PATH_PIN_COLLECTION_SET_OF_3         = "42633"
	CHAMPIONS_PATH_PIN_COLLECTION_SET_OF_3_V2      = "42671"
	CHAMPIONS_PATH_SPECIAL_PIN_COLLECTION_SET_OF_2 = "42672"
	VIVID_VOLTAGE_BB_BOX_PREORDER                  = "43228"
	TEAM_UP_BOOSTER_BOX                            = "33810"
	GALAR_PALS_MINI_TIN_SET_OF_5                   = "40694"
	UNIFIED_MINDS_ETB                              = "41957"
	DARKNESS_ABLAZE_ETB                            = "42000"
	FALL_2020_COLLECTOR_CHEST                      = "43231"
	VIVID_VOLTAGE_ALAKAZAM_V_BOX                   = "44063"

	// DELAYS
	SLEEP_DELAY   = 7200000
	MONITOR_DELAY = 300000
	ERROR_DELAY   = 10000
)

type ProductLink struct {
	sku       string
	name      string
	link      string
	thumbnail string
}

func main() {
	logger, _ := zap.NewProduction()
	defer logger.Sync()

	// TESTING
	// skusToMonitor := []string{
	// 	VIVID_VOLTAGE_BOOSTER_BOX_PREORDER,
	// }

	// LIVE
	skusToMonitor := []string{
		HIDDEN_FATES_ETB,
		KANTO_POWER_COLLECTION,
		CHAMPIONS_PATH_ETB,
		CHAMPIONS_PATH_PIN_COLLECTION_SET_OF_3,
		CHAMPIONS_PATH_PIN_COLLECTION_SET_OF_3_V2,
		CHAMPIONS_PATH_SPECIAL_PIN_COLLECTION_SET_OF_2,
		VIVID_VOLTAGE_BB_BOX_PREORDER,
		TEAM_UP_BOOSTER_BOX,
		GALAR_PALS_MINI_TIN_SET_OF_5,
		UNIFIED_MINDS_ETB,
		DARKNESS_ABLAZE_BOOSTER_BOX,
		VIVID_VOLTAGE_BOOSTER_BOX_PREORDER,
		FALL_2020_COLLECTOR_CHEST,
		VIVID_VOLTAGE_ALAKAZAM_V_BOX,
	}

	productLinkMap := parseProductLinks()

	logger.Info("Parsing proxies...")
	proxies := readProxies()
	logger.Info("Reading proxies successful.")

	var waitGroup sync.WaitGroup
	waitGroup.Add(3)

	i := 0
	p := &i

	for _, sku := range skusToMonitor {
		go checkAvailability(&waitGroup, &proxies, p, sku, productLinkMap, logger)
	}

	waitGroup.Wait()
}

func checkAvailability(wg *sync.WaitGroup, proxies *[]models.Proxy, pos *int, sku string, productMap map[string]ProductLink, logger *zap.Logger) {
	errorDelay := ERROR_DELAY * time.Millisecond

	// Start session, grab cookies
	client := http.Client{}
	err := errors.New("")
	for {
		logger.Info("Initializing session...grabbing cookie...")
		client, err = http.InitializeClient(getProxyByPosition(pos, *proxies), logger)
		if err != nil {
			time.Sleep(errorDelay)
			logger.Info("Could not create session, retrying with new proxy")
			continue
		} else {
			logger.Info("Session created!")

			retryTimer := MONITOR_DELAY * time.Millisecond

			counter := 0
			for {
				logger.Info("Checking product availability...", zap.String("SKU", sku))
				available, product, err := client.GetProductAvailability(sku)
				if err != nil {
					if err.Error() == "invalid character '<' looking for beginning of value" || err.Error() == "invalid character 'A' looking for beginning of value" {
						time.Sleep(errorDelay)
						logger.Info("Could not retrieve availability, reinitializing client")
						break
					} else {
						logger.Error(err.Error(), zap.String("SKU", sku))
						break
					}
				}

				if available {
					counter++
					logger.Info("Product available!", zap.String("SKU", sku))

					p, _ := productMap[sku]
					productName := p.name
					link := p.link
					price := fmt.Sprintf("%.2f", product.Data.Price.WithoutTax.Value)
					thumbnail := p.thumbnail
					stock := fmt.Sprintf("%d", product.Data.Stock)

					foundMessage := "Nurse Joy has found an item in stock on GameNerdz.com!"
					sleepMessage := ""
					if counter == 3 {
						sleepMessage = "\n\n*Item has been found 3 times, Nurse Joy will sleep for 2 hours!*"
					}

					message := models.Message{
						Content: fmt.Sprintf(`%s%s`, foundMessage, sleepMessage),
						Embeds: []models.Embed{
							models.Embed{
								Title:       productName,
								Color:       4437377,
								Description: fmt.Sprintf("<%s>\n\n**SKU:** %s\n**Price:** $%s\n**Stock:** %s", link, sku, price, stock),
								Thumbnail: models.URL{
									URL: fmt.Sprintf("%s", thumbnail),
								},
							},
						},
					}

					retry, err := client.SendDiscordMessage(message)
					if err != nil {
						logger.Error("Error sending discord message", zap.String("errorMessage", err.Error()), zap.String("SKU", sku))
					}
					if retry > MONITOR_DELAY {
						retryTimer, _ = time.ParseDuration(fmt.Sprintf("%dms", retry))
					}

					if counter == 3 {
						sleepTimer := SLEEP_DELAY * time.Millisecond
						logger.Info("Found item 3 times, sleeping....", zap.String("SKU", sku))
						time.Sleep(sleepTimer)
						break
					}
				} else {
					logger.Info("Product unavailable! Retrying....", zap.String("SKU", sku))
					counter = 0
				}

				time.Sleep(retryTimer)
			}
		}
	}

	wg.Done()
}

func readProxies() []models.Proxy {
	proxies := []models.Proxy{}

	file, err := os.Open("./proxies.txt")
	if err != nil {
		return []models.Proxy{}
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		host, un, pw := parseProxyLine(scanner.Text())
		p := models.Proxy{
			Host:   host,
			Un:     un,
			Pw:     pw,
			Status: true,
		}
		proxies = append(proxies, p)
	}
	return proxies
}

func parseProxyLine(proxy string) (string, string, string) {
	parsed := strings.Split(proxy, ":")

	addr, _ := net.LookupIP(parsed[0])
	fmt.Println("IP address: ", addr)

	return strings.Join([]string{addr[rand.Intn(len(addr))].String(), parsed[1]}, ":"), parsed[2], parsed[3]
}

func getValidProxy(proxies *[]models.Proxy) models.Proxy {
	for _, p := range *proxies {
		if p.Status {
			pointer := &p
			*pointer = models.Proxy{
				Status: false,
				Host:   p.Host,
				Un:     p.Un,
				Pw:     p.Pw,
			}

			fmt.Println(strings.Join([]string{p.Host, p.Un, p.Pw}, ":"))

			return p
		}
	}

	resetProxyListStatus(proxies)
	return getValidProxy(proxies)
}

func resetProxyListStatus(proxies *[]models.Proxy) *[]models.Proxy {
	for _, p := range *proxies {
		pointer := &p
		*pointer = models.Proxy{
			Status: true,
			Host:   p.Host,
			Un:     p.Un,
			Pw:     p.Pw,
		}
	}

	return proxies
}

func getProxyByPosition(pos *int, proxies []models.Proxy) models.Proxy {
	if *pos == len(proxies) {
		*pos = 0
	}
	proxy := models.Proxy{
		Status: false,
		Host:   proxies[*pos].Host,
		Un:     proxies[*pos].Un,
		Pw:     proxies[*pos].Pw,
	}

	fmt.Println(strings.Join([]string{proxy.Host, proxy.Un, proxy.Pw}, ":"))

	*pos++
	return proxy
}

func parseProductLinks() map[string]ProductLink {
	csvfile, err := os.Open("./products.csv")
	if err != nil {
		log.Fatalln("Couldn't open the csv file", err)
	}

	r := csv.NewReader(csvfile)
	m := make(map[string]ProductLink)

	for {
		record, err := r.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Fatal(err)
		}

		if record[0] == "sku" {
			continue
		}

		m[record[0]] = ProductLink{
			sku:       record[0],
			name:      record[1],
			link:      record[2],
			thumbnail: record[3],
		}
	}

	return m
}
