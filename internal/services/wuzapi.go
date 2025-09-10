package services

/*
const NAME_DB = "6701d2e3af68f2df279b1091_DB"
const ASSISTANT_ID = "asst_aMlBK71XJFs9qBxNzNttzLer"

const ASSISTANT_CONT = "assistant"

const CREATE_ORDER = "create-order"
const ORDER_STATUS = "order-status"
const HELP = "help"

const MESSAGE_NOT_CLIENT = "¿Eres nuevo por aquí? ¡Ya nos contactaremos contigo!"
const MESSAGE_NOT_UNDERSTAND = "Lo siento, no entendí lo que dijiste. ¿Podrías repetirlo?"

func GetNumberFromWuzapi(data *models.MessageReceiveWuzapi) (string, error) {
	// Obtener el valor de "Sender" y hacer split por "@"
	parts := strings.Split(data.Event.Info.Sender, "@")
	if len(parts) > 0 {
		return parts[0], nil
	}

	return "", validations.ErrInvalidPhone
}

/*
Se busca en la colección de 'clients' si el usuario existe. Utilizando CreateClientIfNotExists().
Agregamos el mensaje al hilo con la función AddMessageToThread().
Recuperamos la respuesta de GPT con la función ListenToGPT().
Le respondemos al usuario utilizando la funcion SendMessageByPhone().
*
func HandleClientThreadAndMessageFlow(numberPhone string, data *models.MessageReceiveWuzapi) {
	SetTypingByPhone(numberPhone)
	//Si el cliente no existe, lo creamos en BD
	client, err := CreateClientIfNotExists(numberPhone, data)
	if err != nil {
		SendMessageByPhone(numberPhone, err.Error())
		return
	}

	//Añadimos el mensaje al hilo
	println("Añadimos mensajes al hilo")
	err = AddMessageToThreadByClient(client, data.Event.Message.Conversation)
	if err != nil {
		SendMessageByPhone(numberPhone, err.Error())
		return
	}
	println("-------------------------")
	println("Escuchamos la respuesta de GPT")
	//Escuchamos la respuesta de GPT
	response, err := ListenToGPT(client.CurrentThread.ID)
	if err != nil {
		SendMessageByPhone(numberPhone, err.Error())
		return
	}

	err = ResponseToUser(client, numberPhone, response)
	if err != nil {
		SendMessageByPhone(numberPhone, err.Error())
		return
	}
	println("-------------------------")
}

/*
Dado un número de telefono y data recibida de wuzapi:
- Busca al cliente en la BD según el Phone.
- Si existe, lo retorna, si no existe, lo crea y lo retorna.
En caso de haber algun fallo, devolvemos un error.
*
func CreateClientIfNotExists(numberPhone string, data *models.MessageReceiveWuzapi) (*models.Client, error) {
	client, err := GetClientByPhone(NAME_DB, numberPhone)
	if err == validations.ErrDocumentNotFound {
		client, err = models.NewClient("", "", data.Event.Info.PushName, numberPhone, "", "")

		if err != nil {
			return nil, validations.ErrCreatingClient
		}

		err = CreateClient(NAME_DB, client)
		if err != nil {
			return nil, validations.ErrCreatingClient
		}
	}
	return client, nil
}

/*
Dado un cliente y el mensaje que nos enviaron por wuzapi:
- Se fija si el cliente tiene el campo thread_id
  - Si no lo tiene, creamos thread y run y updateamos el Client
  - Si lo tiene, solo enviamos el mensaje

En caso de haber algun fallo, devolvemos un error.
*
func AddMessageToThreadByClient(client *models.Client, message string) error {
	//Ahora verificamos que tenga el campo current_thread
	if client.CurrentThread.ID == "" {
		//Creamos una run y el threadID
		threadAndRun, err := CreateThreadAndRun(message, ASSISTANT_ID)
		if err != nil {
			return err
		}

		err = AddThreadIDInClient(NAME_DB, client.ID.Hex(), threadAndRun.ThreadID, client.Threads)
		if err != nil {
			return err
		}

		clientCopy, err := GetClientByID(NAME_DB, client.ID.Hex())
		if err != nil {
			return err
		}

		*client = *clientCopy

	} else {
		//Ya tiene Thread, entonces solo enviamos el mensaje
		//TODO: Si tiene current_thread, debemos validar que tenga menos de 12 horas, sino,
		//lo borramos con una go rutina y hacemos el paso de arriba.

		_, err := SendMessageToThread(client.CurrentThread.ID, message)
		if err != nil {
			return err
		}

		//Corremos la run
		_, err = CreateRun(client.CurrentThread.ID, ASSISTANT_ID)
		if err != nil {
			return err
		}
	}
	return nil
}

/*
Escuchamos la respuestas de GPT con la función GetMessageList.
- Si aún no está la respuesta, volvemos a llamar a la función
- Si dió un error, lo devolvemos.
*
func ListenToGPT(threadID string) (*openai.ResponseGPT, error) {
	messageList, err := GetMessageList(threadID)
	if err != nil {
		return nil, err
	}
	println("Data: ", messageList.Data)
	// Buscamos si el último mensaje (position 0), fue enviado por el asistente y si tiene contenido
	if len(messageList.Data) == 0 {
		return nil, validations.ErrNotMessageOpenAI
	}

	if messageList.Data[0].Role != ASSISTANT_CONT {
		return ListenToGPT(threadID)
	}

	if len(messageList.Data[0].Content) == 0 {
		return ListenToGPT(threadID)
	}

	if messageList.Data[0].Content[0].Text.Value == "" {
		return ListenToGPT(threadID)
	}

	cleanedText := strings.Trim(messageList.Data[0].Content[0].Text.Value, "`")
	println(cleanedText)
	//Retornamos la respuesta de GPT
	var response openai.ResponseGPT
	err = json.Unmarshal([]byte(cleanedText), &response)
	if err != nil {
		return nil, err
	}
	return &response, nil
}

/*
Respondemos al usuario según lo que haya sucedido con ChatGPT
*
func ResponseToUser(client *models.Client, numberPhone string, response *openai.ResponseGPT) error {
	if response.Type == CREATE_ORDER {
		//Si hay Orden y está confirmada, la creamos
		if response.Order != nil && response.Order.IsConfirm {
			_, err := CreateOrderByOrderGPT(NAME_DB, response.Order, client)
			if err != nil {
				return err
			}
			//Respondemos al usuario con la respuesta de la creación de la orden
			SendMessageByPhone(numberPhone, fmt.Sprintf("Tu pedido se ha generado con éxito: \n\n %s \n\n %s", "Detalle del pedido :v", "En caso de algún cambio, no dudes en avisarnos."))
		} else {
			//Respondemos al usuario
			SendMessageByPhone(numberPhone, response.Response)
		}
	} else if response.Type == ORDER_STATUS {
		//Obtenemos resumen de la orden y lo devolvemos
		SendMessageByPhone(numberPhone, fmt.Sprintf("Claro, aquí te enviamos el resumen de tu pedido: \n\n %s", "Resumen del pedido :v"))
	} else if response.Type == HELP {
		//Le avisamos al usuario que ya se contactará un asistente
		SendMessageByPhone(numberPhone, "Ya se contactará un asistente para resolver el problema.")
	} else {
		//TODO: Si es de tipo 'none', respondemos pero agregamos strike.

		//Respondemos al usuario
		SendMessageByPhone(numberPhone, response.Response)
	}
	return nil
}

func SendMessageByPhone(phone string, message string) error {
	go sendMessageToUser(phone, message)
	go SetNotTypingByPhone(phone)
	return nil
}

func SetTypingByPhone(phone string) error {
	go setStatePresence(phone, "composing")
	return nil
}

func SetNotTypingByPhone(phone string) error {
	go setStatePresence(phone, "paused")
	return nil
}

func setStatePresence(phone string, state string) error {
	// Obtener la variable de entorno
	wuzapiURL := os.Getenv("WUZAPI_URL")
	token := os.Getenv("TOKEN_WUZAPI")

	if wuzapiURL == "" {
		return validations.ErrNotTokenWuzapi
	}

	if token == "" {
		return validations.ErrNotTokenWuzapi
	}

	// Crear el cuerpo de la petición
	data := models.SetStateWuzapi{
		Phone: phone,
		State: state,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Crear una nueva solicitud POST
	req, err := http.NewRequest("POST", wuzapiURL+"/chat/presence", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Configurar el header de la solicitud con el token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token) // Agregar el token al header

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	println("Respuesta de Wuzapi:", resp.Status)
	// Verificar la respuesta del servidor
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la petición: %s", resp.Status)
	}

	return nil
}

func sendMessageToUser(phone string, message string) error {
	// Obtener la variable de entorno
	wuzapiURL := os.Getenv("WUZAPI_URL")
	token := os.Getenv("TOKEN_WUZAPI")

	if wuzapiURL == "" {
		return validations.ErrNotTokenWuzapi
	}

	if token == "" {
		return validations.ErrNotTokenWuzapi
	}

	// Crear el cuerpo de la petición
	data := models.SendMessageWuzapi{
		Phone: phone,
		Body:  message,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	// Crear una nueva solicitud POST
	req, err := http.NewRequest("POST", wuzapiURL+"/chat/send/text", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
	}

	// Configurar el header de la solicitud con el token
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("token", token) // Agregar el token al header

	// Enviar la solicitud
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	println("Respuesta de Wuzapi:", resp.Status)
	// Verificar la respuesta del servidor
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("error en la petición: %s", resp.Status)
	}

	return nil
}
*/
