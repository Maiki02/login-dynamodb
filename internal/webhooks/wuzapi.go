package webhook

/*
import (
	"encoding/json"
	"fmt"
	"log"
	"myproject/internal/models"
	"myproject/internal/services"
	"net/http"
)

var usersWithIA = [4]string{
	"5493416887794", //Miqueas
	//"5493415442001", //Antonella
	//"5493416651583", //Martín
	//"5493415312455", //Plácido
	//"5493413544755", //Alana
	//"5493416022527", //Adriel
	//"5493415559634", //Dominguez
	//"5493364336890", //Lucas ID
}

func WebhookWuzapiHandler(w http.ResponseWriter, r *http.Request) {
	// Importar "net/url", "fmt" y "encoding/json"
	if err := r.ParseForm(); err != nil {
		log.Fatal(err)
	}

	// Decodificar jsonData
	jsonData := r.FormValue("jsonData")

	// Definir una estructura para el JSON
	var data models.MessageReceiveWuzapi
	if err := json.Unmarshal([]byte(jsonData), &data); err != nil {
		log.Fatal(err)
	}

	//Solo respondemos si no es de grupo
	if !data.Event.Info.IsGroup {
		fmt.Println("Tipo de evento: ", data.Type)
		AnswerToPhone(&data)
	}

}

func AnswerToPhone(data *models.MessageReceiveWuzapi) {
	if data.Type == "Message" {
		numberPhone, _ := services.GetNumberFromWuzapi(data)

		if searchNumber(numberPhone) {
			services.HandleClientThreadAndMessageFlow(numberPhone, data)
		}

	}
}

func searchNumber(number string) bool {
	for _, v := range usersWithIA {
		if v == number {
			return true
		}
	}
	return false
}
*/
