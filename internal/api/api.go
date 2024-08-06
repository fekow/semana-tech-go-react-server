package api

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"sync"

	"github.com/fekow/semana-tech-go-react-server/internal/store/pgstore"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/jackc/pgx/v5"
)

// devia utilizar uma interface pra poder mudar melhor depois
type apiHandler struct {
	q           *pgstore.Queries
	r           *chi.Mux
	upgrader    websocket.Upgrader
	subscribers map[string]map[*websocket.Conn]context.CancelFunc
	mu          *sync.Mutex
}

//da um upgrade na api pra conseguir lidar com websockets
//MUX É UM MULTIPLEXER

// AQUI É COMO SE CRIA UM METODO PRA ESSA STRUCT
func (h apiHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	h.r.ServeHTTP(w, r)
}

func NewHandler(q *pgstore.Queries) http.Handler {

	// o map nao é thread safe
	// o map inicia como nulo
	// precisamos usar o mutex pra garantir que nao vai dar problema de concorrencia

	a := apiHandler{
		q: q,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool { return true },
			// deixa true so em dev
		},
		subscribers: make(map[string]map[*websocket.Conn]context.CancelFunc),
		mu:          &sync.Mutex{},
	}

	r := chi.NewRouter()

	r.Use(middleware.Logger, middleware.Recoverer, middleware.RequestID)

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins:   []string{"https://*", "http://*"},
		AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type", "X-CSRF-Token"},
		ExposedHeaders:   []string{"Link"},
		AllowCredentials: false,
		MaxAge:           300,
	}))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Hello, World!"))
	})

	r.Get("/subscribe/{room_id}", a.handleSubscribe)

	r.Route("/api", func(r chi.Router) {
		r.Route("/rooms", func(r chi.Router) {
			r.Post("/", a.handleCreateRoom)
			r.Get("/", a.handleGetRoom)
			r.Route("/{room_id}/messages", func(r chi.Router) {

				r.Get("/", a.handleGetRoomMessages)
				r.Post("/", a.handleCreateRoomMessage)
				r.Route("/{message_id}", func(r chi.Router) {
					r.Get("/", a.handleGetRoomMessages)
					r.Patch("/react", a.handleReactToMessage)
					r.Delete("/react", a.handleRemoveReactFromMessage)
					r.Patch("/answer", a.handleMarkMessageAsAnswered)
				})
			})
		})
	})

	a.r = r
	return a
}

func (h apiHandler) handleSubscribe(w http.ResponseWriter, r *http.Request) {

	rawRoomId := chi.URLParam(r, "room_id")

	slog.Info("new client tried to connect", "room_id", rawRoomId, "client_ip", r.RemoteAddr)

	roomId, err := uuid.Parse(rawRoomId)
	if err != nil {
		http.Error(w, "invalid room id", http.StatusBadRequest)
		return
	}

	_, err = h.q.GetRoom(r.Context(), roomId)
	if err != nil {

		if errors.Is(err, pgx.ErrNoRows) {
			http.Error(w, "room not found", http.StatusNotFound)
			return
		}
		http.Error(w, "something went wrong", http.StatusInternalServerError)
		return
	}

	c, err := h.upgrader.Upgrade(w, r, nil)

	if err != nil {
		slog.Warn("failed to upgrade connection", "error", err)
		http.Error(w, "could not upgrade to ws connection", http.StatusBadRequest)
		return
	}

	// limpeza de recursos
	defer c.Close()

	ctx, cancel := context.WithCancel(r.Context())

	h.mu.Lock()
	if _, ok := h.subscribers[rawRoomId]; !ok {
		h.subscribers[rawRoomId] = make(map[*websocket.Conn]context.CancelFunc)
	}
	slog.Info("new client connected", "room_id", rawRoomId, "client_ip", r.RemoteAddr)
	h.subscribers[rawRoomId][c] = cancel

	h.mu.Unlock()

	// 	se o cliente ou o servidor cancelar a conexao, a funcao cancel ele cai no done aqui
	<-ctx.Done()

	h.mu.Lock()
	delete(h.subscribers[rawRoomId], c)
	h.mu.Unlock()

	// for {
	// 	_, msg, err := conn.ReadMessage()
	// 	if err != nil {
	// 		break
	// 	}

	// 	h.q.SendMessage(r.Context(), pgstore.SendMessageParams{
	// 		RoomID: room_id,
	// 		Message: string(msg),
	// 	})
	// }

}

func (h apiHandler) handleCreateRoom(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleGetRoom(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleGetRoomMessages(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleGetRoomMessage(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleCreateRoomMessage(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleReactToMessage(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleRemoveReactFromMessage(w http.ResponseWriter, r *http.Request) {

}
func (h apiHandler) handleMarkMessageAsAnswered(w http.ResponseWriter, r *http.Request) {

}
