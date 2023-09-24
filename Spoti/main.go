package main

import (
	"encoding/json" //Librería para decodificar el mp3 en json para poder leer nombres de canciones
	"fmt"
	"github.com/faiface/beep" //Libreería beep para la reproducción de la música
	"github.com/faiface/beep/mp3"
	"github.com/faiface/beep/speaker"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Canciones con esta estructura
type Cancion struct {
	Nombre string
}

// Playlist con esta estructura
type Playlist struct {
	Nombre    string
	Canciones []Cancion
}

var musicFolder = "/Users/rsanchez/Documents/MusicProject" //Variable donde se almacena la ruta de las canciones que hay en local
var playlists []Playlist                                   //Las Playlis se alamacenan en variables

func main() {
	startHTTPServer() //Inicio de la función para el protocolo http
}

// CORS para la parte web
func enableCors(w *http.ResponseWriter) {
	(*w).Header().Set("Access-Control-Allow-Origin", "*")
	(*w).Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
	(*w).Header().Set("Access-Control-Allow-Headers", "Accept, Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization")
}

// Función para poder conectarnos desde otra aplicación o cliente.
func startHTTPServer() {
	httpPort := ":8081" //Estamos en el puerto 8081
	fmt.Println("Servidor HTTP en ejecución en el puerto", httpPort)

	//endpoints
	http.HandleFunc("/ver", verCancionesHandler)                                //ve todas las canciones
	http.HandleFunc("/reproducir", reproducirCancionHandler)                    //reproduce las canciones
	http.HandleFunc("/crearPlaylist", crearPlaylistHandler)                     //crea playlists
	http.HandleFunc("/agregarCancionAPlaylist", agregarCancionAPlaylistHandler) //agrega canciones a playlists
	//Si hay error en la conexión
	if err := http.ListenAndServe(httpPort, nil); err != nil {
		fmt.Println("Error al iniciar el servidor HTTP:", err)
	}
}

// Función para que el endpoint /ver funcione.
func verCancionesHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Println("Nueva conexión HTTP: /ver, desde", r.RemoteAddr) // Log para mostrar nueva conexión

	fileInfos, err := os.ReadDir(musicFolder)
	if err != nil {
		http.Error(w, "Error al leer el directorio de música", http.StatusInternalServerError)
		return
	}

	var canciones []Cancion
	for _, fileInfo := range fileInfos {
		if !fileInfo.IsDir() && strings.HasSuffix(fileInfo.Name(), ".mp3") {
			canciones = append(canciones, Cancion{Nombre: fileInfo.Name()})
		}
	}
	//Una excepción si no se obtiene como json las cacniones.
	jsonResponse, err := json.Marshal(canciones)
	if err != nil {
		http.Error(w, "Error al convertir los nombres de las canciones a JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonResponse)
}

// Función para que el endpoint /reproducir funcione.
func reproducirCancionHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Println("Nueva conexión HTTP: /reproducir, desde", r.RemoteAddr) // Log para mostrar nueva conexión

	query := r.URL.Query().Get("nombre")
	if query == "" {
		http.Error(w, "Parámetro 'nombre' es requerido", http.StatusBadRequest)
		return
	}

	rutaCompleta := filepath.Join(musicFolder, query)
	//Diferentes erorres por si existe un error en la reproducción de la canción.
	file, err := os.Open(rutaCompleta)
	if err != nil {
		http.Error(w, "Error al abrir el archivo de la canción", http.StatusInternalServerError)
		return
	}
	defer file.Close()

	streamer, format, err := mp3.Decode(file)
	if err != nil {
		http.Error(w, "Error al decodificar el archivo MP3", http.StatusInternalServerError)
		return
	}
	defer streamer.Close()

	err = speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10))
	if err != nil {
		http.Error(w, "Error al inicializar el speaker", http.StatusInternalServerError)
		return
	}
	//Usa libería beep para poder utilizar funciones de reproducir la música.
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))

	<-done

	fmt.Fprintf(w, "Canción %s reproducida con éxito", query)
}

// Función para que el endpoint /crearPlaylist funcione.
func crearPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Println("Nueva conexión HTTP: /crearPlaylist, desde", r.RemoteAddr)

	var nuevaPlaylist Playlist
	err := json.NewDecoder(r.Body).Decode(&nuevaPlaylist)
	if err != nil {
		http.Error(w, "Error al leer el cuerpo de la solicitud", http.StatusBadRequest)
		return
	}

	playlists = append(playlists, nuevaPlaylist)
	fmt.Fprintf(w, "Playlist %s creada con éxito", nuevaPlaylist.Nombre)
}

// Función para que el endpoint /añadirCancionAPlaylist funcione.
func agregarCancionAPlaylistHandler(w http.ResponseWriter, r *http.Request) {
	enableCors(&w)
	log.Println("Nueva conexión HTTP: /agregarCancionAPlaylist, desde", r.RemoteAddr)
}
