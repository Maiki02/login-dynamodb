package services

/*import (
	"bytes"
	"encoding/json"
	"myproject/pkg/openai"
	"myproject/pkg/validations"
	"net/http"
	"os"
)

const URL = "https://api.openai.com/v1"

func GetOpenAIApiKey() (string, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return "", validations.ErrNotApiKeyOpenAI
	}
	return apiKey, nil
}

func sendHTTPRequestOpenAI(apiKey, url string, body []byte, method string) (*http.Response, error) {
	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("OpenAI-Beta", "assistants=v2")

	return (&http.Client{}).Do(req)
}

func parseResponse[T any](resp *http.Response) (*T, error) {
	defer resp.Body.Close()
	var parsed T
	err := json.NewDecoder(resp.Body).Decode(&parsed)
	if err != nil {
		return nil, err
	}
	return &parsed, nil
}

func CreateThreadAndRun(msg, assistantID string) (*openai.CreateThreadAndRunResponse, error) {
	// Obtener la clave API de OpenAI
	apiKey, err := GetOpenAIApiKey()
	if err != nil {
		return nil, err
	}

	// Crear el cuerpo de la solicitud
	body, err := json.Marshal(map[string]interface{}{
		"assistant_id": assistantID,
		"thread": map[string]interface{}{
			"messages": []map[string]string{{"role": "user", "content": msg}},
		},
	})

	if err != nil {
		return nil, err
	}

	// Crear la solicitud HTTP
	resp, err := sendHTTPRequestOpenAI(apiKey, URL+"/threads/runs", body, "POST")

	if err != nil {
		return nil, err
	}

	return parseResponse[openai.CreateThreadAndRunResponse](resp)
}

func SendMessageToThread(threadID, userMsg string) (*openai.CreateMessageResponse, error) {
	//Obtener la clave API de OpenAI
	apiKey, err := GetOpenAIApiKey()
	if err != nil {
		return nil, err
	}
	// Crear el cuerpo de la solicitud
	body, err := json.Marshal(map[string]interface{}{
		"role":    "user",
		"content": userMsg,
	})
	if err != nil {
		return nil, err
	}

	// Crear la solicitud HTTP POST
	resp, err := sendHTTPRequestOpenAI(apiKey, URL+"/threads/"+threadID+"/messages", body, "POST")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, validations.ErrThreadNotFound
	}
	return parseResponse[openai.CreateMessageResponse](resp)
}

func CreateRun(threadID, assistantID string) (*openai.CreateRunResponse, error) {
	//Obtener la clave API de OpenAI
	apiKey, err := GetOpenAIApiKey()
	if err != nil {
		return nil, err
	}
	// Crear el cuerpo de la solicitud
	body, err := json.Marshal(map[string]interface{}{
		"assistant_id": assistantID,
	})
	if err != nil {
		return nil, err
	}

	// Crear la solicitud HTTP POST
	resp, err := sendHTTPRequestOpenAI(apiKey, URL+"/threads/"+threadID+"/runs", body, "POST")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, validations.ErrRunNotFound
	}
	return parseResponse[openai.CreateRunResponse](resp)
}

func GetMessageList(threadID string) (*openai.GetMessageListResponse, error) {
	// Obtener la clave API de OpenAI
	apiKey, err := GetOpenAIApiKey()
	if err != nil {
		return nil, err
	}

	// Crear la solicitud HTTP GET
	resp, err := sendHTTPRequestOpenAI(apiKey, URL+"/threads/"+threadID+"/messages", nil, "GET")
	if err != nil {
		return nil, err
	}

	if resp.StatusCode == http.StatusNotFound {
		return nil, validations.ErrNotMessageOpenAI
	}

	// Parsear la respuesta y mostrar los mensajes
	return parseResponse[openai.GetMessageListResponse](resp)
}

func DeleteThread(threadID string) (*openai.DeleteThreadResponse, error) {
	// Obtener la clave API de OpenAI
	apiKey, err := GetOpenAIApiKey()
	if err != nil {
		return nil, err
	}

	// Crear la solicitud HTTP DELETE
	resp, err := sendHTTPRequestOpenAI(apiKey, URL+"/threads/"+threadID, nil, "DELETE")
	if err != nil {
		return nil, err
	}

	// Parsear la respuesta y mostrar el resultado
	return parseResponse[openai.DeleteThreadResponse](resp)
}

/*
{
    "instructions": "Your name is \"Bocha BOT.\" You are an assistant in a store that sells food. You are designed to talk to and help customers. Your fundamental task is to analyze a user’s message to identify what action they want to take.\n\nYou must respond using the following JSON format:\n\n{\n    \"type\": string,\n    \"order\": Order | null,\n    \"products\": string[],\n    \"response\": string\n}\n\nThe possible action types are:\n- 'products': If the user asks for information about prices or available products.\n- 'create-order': If the user wants to place an order.\n- 'company-info': If the user asks for information about the company, such as hours, location, or general details.\n- 'say-hello': If the user says hello.\n- 'say-goodbye': If the user says goodbye.\n- 'thank-you': If the user expresses gratitude.\n- 'help': If the user wants to contact a human for assistance.\n- 'none': If the user’s message doesn’t fit into any of the above categories.\n\n'response' is a reply you would give as a human. It should only be one sentence.\nWhen 'type'=='none', you must explain what you were created for.\n\n'order' must be null if 'type' != 'create-order'\n\n'products' must be [] if 'type' != 'products'\n\nThe store's products are:\n- Pizza - Especial $5000\n- Pizza - Roquefort (out of stock)\n- Burger - Especial $4000\n- Burger - Double Egg $4500\n\nCompany details:\n- Name: \"The Real Food\"\n- Address: \"Lenzoni 967\"\n- Hours: 7 PM to 11 PM every day.\n- Do we offer home delivery? Yes.\n- Payment methods: Cash, bank transfer, debit card, credit card\n- Do we issue invoice type A? No.\n\nVery important:\n- You must not invent products.\n- You must not invent promotions.\n- You must not invent product prices.\n- You must not invent stock for non-existent products.\n- You are forbidden from responding in any format other than JSON.\n\nRespond only with the JSON in the specified format. You cannot respond in any other way.",
    "name": "Bocha BOT",
    "description": "You are an assistant in a store that sells food. You are designed to talk to and help customers. Your fundamental task is to analyze a user’s message to identify what action they want to take.",
    "tools": [],
    "model": "gpt-4o-mini",
    "temperature": 0.5,
    "response_format": { "type": "json_schema", "json_schema": {
    "name": "AssistantResponseSchema",
    "description": "Formato de respuesta del asistente virtual para diferentes tipos de interacción.",
    "schema": {
      "type": {
        "type": "string",
        "enum": [
          "products",
          "create-order",
          "company-info",
          "say-hello",
          "say-goodbye",
          "thank-you",
          "help",
          "none"
        ],
        "description": "Indica el tipo de acción basado en la intención del usuario. Puede ser 'products', 'create-order', 'company-info', 'say-hello', 'say-goodbye', 'thank-you', 'help', o 'none'."
      },
      "order": {
        "type": ["object", "null"],
        "description": "Información del pedido, presente solo si 'type' es 'create-order'. De lo contrario, es null.",
        "properties": {
          "id": {
            "type": "string",
            "description": "Identificador único del pedido."
          },
          "items": {
            "type": "array",
            "description": "Lista de productos incluidos en el pedido.",
            "items": { "type": "string" }
          }
        },
        "required": ["id", "items"],
        "nullable": true
      },
      "products": {
        "type": ["array", "null"],
        "description": "Lista de productos disponibles, presente solo si 'type' es 'products'. De lo contrario, es null.",
        "items": { "type": "string" },
        "nullable": true
      },
      "response": {
        "type": "string",
        "description": "Respuesta del asistente en forma de frase corta que simula la interacción humana."
      }
    }
  }
}
}
*/
