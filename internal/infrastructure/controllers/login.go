package controllers

import (
	"encoding/json"
	"fmt"
	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type DiscordUser struct {
	ID            string `json:"id"`
	Username      string `json:"username"`
	Discriminator string `json:"discriminator"`
	Avatar        string `json:"avatar"`
}

// LoginHandler Redirige l'utilisateur vers Discord pour l'authentification
func (ctx *AppContext) LoginHandler(c *gin.Context) {
	authURL := fmt.Sprintf(
		"https://discord.com/api/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&scope=identify%%20guilds%%20guilds.members.read",
		strconv.Itoa(ctx.Config.Discord.ClientID), ctx.Config.Server.PublicUrl+"/callback")
	c.Redirect(http.StatusTemporaryRedirect, authURL)
}

// CallbackHandler Gère le callback Discord et vérifie les autorisations
func (ctx *AppContext) CallbackHandler(c *gin.Context) {
	code := c.Query("code")
	if code == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "No code provided"})
		return
	}

	// Échanger le code contre un jeton
	tokenResp, err := http.PostForm("https://discord.com/api/oauth2/token", map[string][]string{
		"client_id":     {strconv.Itoa(ctx.Config.Discord.ClientID)},
		"client_secret": {ctx.Config.Discord.Secret},
		"grant_type":    {"authorization_code"},
		"code":          {code},
		"redirect_uri":  {ctx.Config.Server.PublicUrl + "/callback"},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch token"})
		return
	}
	defer tokenResp.Body.Close()

	var tokenData map[string]interface{}
	json.NewDecoder(tokenResp.Body).Decode(&tokenData)

	accessToken, ok := tokenData["access_token"].(string)
	if !ok {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid token response"})
		return
	}

	// Récupération de l'ID utilisateur
	userID, err := getUserID(accessToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err})
	}

	// Stocker l'état d'authentification dans la session
	session := sessions.Default(c)
	session.Set("authenticated", true)
	session.Set("userID", userID)
	session.Save()
	c.Redirect(http.StatusTemporaryRedirect, ctx.Config.Server.PublicUrl+"/game/test")
}

func getUserID(accessToken string) (string, error) {
	// URL de l'API Discord pour récupérer les informations de l'utilisateur
	url := "https://discord.com/api/users/@me"

	// Création d'une requête HTTP
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", fmt.Errorf("erreur lors de la création de la requête: %v", err)
	}

	// Ajout du token d'accès dans l'en-tête Authorization
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Envoi de la requête
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("erreur lors de l'exécution de la requête: %v", err)
	}
	defer resp.Body.Close()

	// Vérification du code de statut HTTP
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("erreur de réponse HTTP, code: %d", resp.StatusCode)
	}

	// Décodage de la réponse JSON
	var user DiscordUser
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return "", fmt.Errorf("erreur lors du décodage de la réponse JSON: %v", err)
	}

	// Retourner l'ID utilisateur
	return user.ID, nil
}
