package ws

type Message struct {
	Room string
	Data []byte
	// Binary true bo'lsa WS binary frame sifatida yuboriladi (collab op'lar —
	// msgpack+deflate). Backend event'lari (JSON) Binary=false (text frame).
	Binary bool
	// Sender — client-originated xabar manbasi. Hub uni o'ziga qaytarmaydi
	// (echo yo'q). Backend (Redis) xabarlarida nil — hammaga yuboriladi.
	Sender *Client
}
