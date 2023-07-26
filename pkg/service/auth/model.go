package auth

/**
(map[string]interface {}) (len=13) {
 (string) (len=10) `json:"given_name"`: (string) (len=6) `json:"Alfred"`,
 (string) (len=7) `json:"picture"`: (string) (len=93) `json:"https://lh3.googleusercontent.com/a/AAcHTtfL0bOOxHvAiZ0BooMD0JuOQ8e2u6tNCyZBztW2Uf5nupQ=s96-c"`,
 (string) (len=3) `json:"iss"`: (string) (len=29) `json:"https://avalond.eu.auth0.com/"`,
 (string) (len=3) `json:"aud"`: (string) (len=32) `json:"l6D6iSi8M3rV9fvnNOuyacKGjGLMFx0y"`,
 (string) (len=3) `json:"iat"`: (float64) 1.689677622e+09,
 (string) (len=3) `json:"sub"`: (string) (len=35) `json:"google-oauth2|109426497729194414452"`,
 (string) (len=11) `json:"family_name"`: (string) (len=7) `json:"Dobradi"`,
 (string) (len=8) `json:"nickname"`: (string) (len=13) `json:"alfreddobradi"`,
 (string) (len=4) `json:"name"`: (string) (len=14) `json:"Alfred Dobradi"`,
 (string) (len=6) `json:"locale"`: (string) (len=2) `json:"en"`,
 (string) (len=10) `json:"updated_at"`: (string) (len=24) `json:"2023-07-18T09:02:06.322Z"`,
 (string) (len=3) `json:"exp"`: (float64) 1.689713622e+09,
 (string) (len=3) `json:"sid"`: (string) (len=32) `json:"CcC6x6lUV6V11lvjC0J67tNVVcAnY2Op"`
}
(map[string]interface {}) (len=16) {
 (string) (len=10) `json:"updated_at"`: (string) (len=24) `json:"2023-07-18T09:02:06.322Z"`,
 (string) (len=7) `json:"user_id"`: (string) (len=35) `json:"google-oauth2|109426497729194414452"`,
 (string) (len=10) `json:"given_name"`: (string) (len=6) `json:"Alfred"`,
 (string) (len=7) `json:"picture"`: (string) (len=93) `json:"https://lh3.googleusercontent.com/a/AAcHTtfL0bOOxHvAiZ0BooMD0JuOQ8e2u6tNCyZBztW2Uf5nupQ=s96-c"`,
 (string) (len=12) `json:"app_metadata"`: (map[string]interface {}) (len=1) {
  (string) (len=11) `json:"external_id"`: (string) (len=36) `json:"ae84697c-68d5-43d6-86bb-71fd96a9ab5a"`
 },
 (string) (len=10) `json:"last_login"`: (string) (len=24) `json:"2023-07-18T07:29:55.521Z"`,
 (string) (len=12) `json:"logins_count"`: (float64) 1,
 (string) (len=10) `json:"identities"`: ([]interface {}) (len=1 cap=1) {
  (map[string]interface {}) (len=4) {
   (string) (len=8) `json:"provider"`: (string) (len=13) `json:"google-oauth2"`,
   (string) (len=7) `json:"user_id"`: (string) (len=21) `json:"109426497729194414452"`,
   (string) (len=10) `json:"connection"`: (string) (len=13) `json:"google-oauth2"`,
   (string) (len=8) `json:"isSocial"`: (bool) true
  }
 },
 (string) (len=8) `json:"nickname"`: (string) (len=13) `json:"alfreddobradi"`,
 (string) (len=4) `json:"name"`: (string) (len=14) `json:"Alfred Dobradi"`,
 (string) (len=7) `json:"last_ip"`: (string) (len=13) `json:"84.203.97.221"`,
 (string) (len=10) `json:"created_at"`: (string) (len=24) `json:"2023-07-18T07:29:55.523Z"`,
 (string) (len=6) `json:"locale"`: (string) (len=2) `json:"en"`,
 (string) (len=11) `json:"family_name"`: (string) (len=7) `json:"Dobradi"`,
 (string) (len=5) `json:"email"`: (string) (len=23) `json:"alfreddobradi@gmail.com"`,
 (string) (len=14) `json:"email_verified"`: (bool) true
}
**/

type Profile struct {
	GivenName  string  `json:"given_name"`
	Picture    string  `json:"picture"`
	Issuer     string  `json:"iss"`
	Audience   string  `json:"aud"`
	IssuedAt   float64 `json:"iat"`
	Subject    string  `json:"sub"`
	FamilyName string  `json:"family_name"`
	Nickname   string  `json:"nickname"`
	Name       string  `json:"name"`
	Locale     string  `json:"locale"`
	UpdatedAt  string  `json:"updated_at"`
	ExpiresAt  float64 `json:"exp"`
	SessionID  string  `json:"sid"`
}

type UserProfile struct {
	AppMetadata   map[string]interface{} `json:"app_metadata"`
	UserMetadata  map[string]interface{} `json:"user_metadata"`
	Email         string                 `json:"email"`
	EmailVerified bool                   `json:"email_verified"`
}
