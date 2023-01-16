# SfcSourceFlowBot Environment Variable #
This envar is used to represents the different options or reason for which the chat was requested. Which this map we can modify the different queues according to the client's needs.

Option struct:

```json
flow={
  "subject": "subjet",
  "providers": {
    "whatsapp": {
      "button_id": "button_wa_id",
      "owner_id": "owner_wa_id"
    },
    "facebook": {
      "button_id": "button_fb_id",
      "owner_id": "owner_fb_id"
    }
  }
}
...
```

If the client only has a single button configured, it would be as follows:

```json
flow={
  "subject": "subjet",
  "providers": {
    "whatsapp": {
      "button_id": "button_id",
      "owner_id": "owner_id"
    },
    "facebook": {
      "button_id": "button_id",
      "owner_id": "owner_id"
    }
  }
}
...
```

Here we have a more complete example according to coppel's needs:

**SFB001. Quiero saber dónde está mi pedido** -> Cola de Atención

**SFB002. Quiero encontrar o comprar un artículo** -> Cola de Venta asistida

**SFB003. Quiero mi Estado de Cuenta** -> Cola de Atención

**SFB004. Quiero registrar o saber el estatus de mi queja** -> Cola de Atención

**SFB005. Solicitar cancelación de mi pedido** - > Cola de Devoluciones

**SFB006. Solicitar servicio de garantía** - > Campaña Chat: Cola: Solicitar servicio de garantía

**SFB007. Solicitar información sobre mi abono ->** Cola de Abonos

**SFB008. Recibí una notificación y deseo más información -> Cola de** Atención

**SFB009. Alguna otra duda - >** Cola de Atención

```json
SFB001={
	"subject":"Quiero saber dónde está mi pedido",
	"providers":{
		"whatsapp":{
			"button_id":"5737**********",
			"owner_id":"00G7***********"
		},
		"facebook":{
			"button_id":"5737***********",
			"owner_id":"00G7***********"
		}
	}
};
SFB002={
	"subject":"Quiero encontrar o comprar un artículo",
	"providers":{
		"whatsapp":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		},
		"facebook":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		}
	}
};
SFB003={
	"subject":"Quiero mi Estado de Cuenta",
	"providers":{
		"whatsapp":{
			"button_id":"5737**********",
			"owner_id":"00G7***********"
		},
		"facebook":{
			"button_id":"5737***********",
			"owner_id":"00G7***********"
		}
	}
};
SFB004={
	"subject":"Quiero registrar o saber el estatus de mi queja",
	"providers":{
		"whatsapp":{
			"button_id":"5737**********",
			"owner_id":"00G7***********"
		},
		"facebook":{
			"button_id":"5737***********",
			"owner_id":"00G7***********"
		}
	}
};
SFB005={
	"subject":"Solicitar cancelación de mi pedido",
	"providers":{
		"whatsapp":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		},
		"facebook":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		}
	}
};
SFB006={
	"subject":"Solicitar servicio de garantía",
	"providers":{
		"whatsapp":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		},
		"facebook":{
			"button_id":"5737b*************",
			"owner_id":"00G7b*************"
		}
	}
};
SFB007={
	"subject":"Solicitar información sobre mi abono",
	"providers":{
		"whatsapp":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		},
		"facebook":{
			"button_id":"5737b**********",
			"owner_id":"00G7b*************"
		}
	}
};
SFB008={
	"subject":"Recibí una notificación y deseo más información",
	"providers":{
			"whatsapp":{
				"button_id":"5737**********",
				"owner_id":"00G7***********"
			},
			"facebook":{
				"button_id":"5737***********",
				"owner_id":"00G7***********"
			}
	}
};
SFB009={
	"subject":" Alguna otra duda",
	"providers":{
		"whatsapp":{
			"button_id":"5737**********",
			"owner_id":"00G7***********"
		},
		"facebook":{
			"button_id":"5737***********",
			"owner_id":"00G7***********"
		}
	}
};
default={
	"subject":" Alguna otra duda",
	"providers":{
		"whatsapp":{
			"button_id":"5737**********",
			"owner_id":"00G7***********"
		},
		"facebook":{
			"button_id":"5737***********",
			"owner_id":"00G7***********"
		}
	}
}
```
