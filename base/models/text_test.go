package models

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestMessageTemplate_Decode(t *testing.T) {
	type args struct {
		value string
	}
	tests := []struct {
		name    string
		sd      *MessageTemplate
		args    args
		wantErr bool
		want    *MessageTemplate
	}{
		{
			name: "Decode success",
			sd:   &MessageTemplate{},
			args: args{
				value: `{"waitAgent":"Esperando un agente","welcomeTemplate":"Hola soy %s y necesito ayuda","context":"Contexto","DescriptionCase":"caso levantado por el Bot","uploadImageError":"Imagen no enviada","uploadImageSuccess":"**El usuario adjunto una imagen al caso**","queuePosition":"Posición en la cola","waitTime":"Tiempo de espera"}`,
			},
			wantErr: false,
			want: &MessageTemplate{
				WaitAgent:          "Esperando un agente",
				QueuePosition:      "Posición en la cola",
				WaitTime:           "Tiempo de espera",
				WelcomeTemplate:    "Hola soy %s y necesito ayuda",
				Context:            "Contexto",
				DescriptionCase:    "caso levantado por el Bot",
				UploadImageError:   "Imagen no enviada",
				UploadImageSuccess: "**El usuario adjunto una imagen al caso**",
			},
		},
		{
			name: "Decode Fail",
			sd:   &MessageTemplate{},
			args: args{
				value: `"waitAgent":"Esperando un agente","welcomeTemplate":"Hola soy %s y necesito ayuda","context":"Contexto","DescriptionCase":"caso levantado por el Bot","uploadImageError":"Imagen no enviada","uploadImageSuccess":"**El usuario adjunto una imagen al caso**","queuePosition":"Posición en la cola","waitTime":"Tiempo de espera"}`,
			},
			wantErr: true,
			want: &MessageTemplate{
				WaitAgent:          "Esperando un agente",
				QueuePosition:      "Posición en la cola",
				WaitTime:           "Tiempo de espera",
				WelcomeTemplate:    "Hola soy %s y necesito ayuda",
				Context:            "Contexto",
				DescriptionCase:    "caso levantado por el Bot",
				UploadImageError:   "Imagen no enviada",
				UploadImageSuccess: "**El usuario adjunto una imagen al caso**",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.sd.Decode(tt.args.value)
			if (err != nil) != tt.wantErr {
				t.Errorf("SfcSourceFlowBot.Decode() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr == true {
				assert.Error(t, err)
			} else {
				assert.Equal(t, tt.want, tt.sd)
			}
		})
	}
}
