package provider

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/resource"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func TestDataSource_200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`

								terraform {
									required_providers {
						  			http = {
										source = "MehdiAtBud/http"
										version ="2.2.7"
						  			}
								}
				  			}
							data "http" "http_test" {
								url = "%s/200"
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", "1.0.0"),
					resource.TestCheckResourceAttr("data.http.http_test", "response_headers.Content-Type", "text/plain"),
					resource.TestCheckResourceAttr("data.http.http_test", "response_headers.X-Single", "foobar"),
					resource.TestCheckResourceAttr("data.http.http_test", "response_headers.X-Double", "1, 2"),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "200"),
				),
			},
		},
	})
}

func TestDataSource_404(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
								url = "%s/404"
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", ""),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "404"),
				),
			},
		},
	})
}

func TestDataSource_withAuthorizationRequestHeader_200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
								url = "%s/restricted"

								request_headers = {
									"Authorization" = "Zm9vOmJhcg=="
								}
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", "1.0.0"),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "200"),
				),
			},
		},
	})
}

func TestDataSource_withAuthorizationRequestHeader_403(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
  								url = "%s/restricted"

  								request_headers = {
    								"Authorization" = "unauthorized"
  								}
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", ""),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "403"),
				),
			},
		},
	})
}

func TestDataSource_utf8_200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ProviderFactories: providerFactories,
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
  								url = "%s/utf-8/200"
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", "1.0.0"),
					resource.TestCheckResourceAttr("data.http.http_test", "response_headers.Content-Type", "text/plain; charset=UTF-8"),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "200"),
				),
			},
		},
	})
}

func TestDataSource_utf16_200(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
  								url = "%s/utf-16/200"
							}`, testHttpMock.server.URL),
				// This should now be a warning, but unsure how to test for it...
				// ExpectWarning: regexp.MustCompile("Content-Type is not a text type. Got: application/json; charset=UTF-16"),
			},
		},
	})
}

func TestDataSource_x509cert(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.7",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: fmt.Sprintf(`
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.7"
					  				}
								}
				  			}
							data "http" "http_test" {
  								url = "%s/x509-ca-cert/200"
							}`, testHttpMock.server.URL),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "response_body", "pem"),
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "200"),
				),
			},
		},
	})
}

func TestDataSource_NonRegisteredDomainBackoff(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.16",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
						
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.16"
					  				}
								}
				  			}
							data "http" "http_test" {
								url = "https://non-existing.thisisbud.com"
								max_elapsed_time = 10
								initial_interval = 100
								multiplier = 1.2
								max_interval = 5000
							}`,
				Check:       resource.ComposeTestCheckFunc(),
				ExpectError: regexp.MustCompile("no such host"),
			},
		},
	})
}

func TestDataSource_RegisteredDomainBackoff(t *testing.T) {
	testHttpMock := setUpMockHttpServer()
	defer testHttpMock.server.Close()

	resource.UnitTest(t, resource.TestCase{
		Providers: map[string]*schema.Provider{
			"http-wait": New("dev")(),
		},
		ExternalProviders: map[string]resource.ExternalProvider{
			"http": {
				VersionConstraint: "2.2.16",
				Source:            "MehdiAtBud/http",
			},
		},
		Steps: []resource.TestStep{
			{
				Config: `
							terraform {
								required_providers {
					  				http = {
									source = "MehdiAtBud/http"
									version ="2.2.16"
					  				}
								}
				  			}

							data "http" "http_test" {
								url = "https://example.com"
								max_elapsed_time = 10
								initial_interval = 100
								multiplier = 1.2
								max_interval = 5000
							}`,
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("data.http.http_test", "status_code", "200"),
				),
			},
		},
	})
}

type TestHttpMock struct {
	server *httptest.Server
}

func setUpMockHttpServer() *TestHttpMock {
	Server := httptest.NewServer(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Add("X-Single", "foobar")
			w.Header().Add("X-Double", "1")
			w.Header().Add("X-Double", "2")

			switch r.URL.Path {
			case "/200":
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/restricted":
				if r.Header.Get("Authorization") == "Zm9vOmJhcg==" {
					w.WriteHeader(http.StatusOK)
					_, _ = w.Write([]byte("1.0.0"))
				} else {
					w.WriteHeader(http.StatusForbidden)
				}
			case "/utf-8/200":
				w.Header().Set("Content-Type", "text/plain; charset=UTF-8")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/utf-16/200":
				w.Header().Set("Content-Type", "application/json; charset=UTF-16")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("1.0.0"))
			case "/x509-ca-cert/200":
				w.Header().Set("Content-Type", "application/x-x509-ca-cert")
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("pem"))
			default:
				w.WriteHeader(http.StatusNotFound)
			}
		}),
	)

	return &TestHttpMock{
		server: Server,
	}
}
