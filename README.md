# live-quest

![Gopher](assets/animations/gopher/standing/gopher_standing.gif)

Este é o nosso primeiro projeto na Codigolândia: um jogo desenvolvido em **Go**
e que pode ser jogado durante as lives!

Neste jogo cada participante terá seu avatar exibido como um Gopher na live,
e as interações no chat poderão ser utilizadas para movimentar o seu
personagem, atacar outros jogadores e muito mais!

Como streammer, você poderá visualizar os Gophers e também ter uma visão
consolidada do histórico do Chat das plataformas suportadas.


## Como funciona?

O programa deve ser executado antes de entrar ao vivo. Ele funcionará
como uma janela de Jogo, que poderá ser capturada e sobreposta no OBS
como um novo Source.

### Instalação

Para instalar o LiveQuest, basta baixar na área de
[Releases](https://github.com/codigolandia/live-quest/releases) e escolher
o arquivo apropriado para o seu sistema operacional. Após baixar, basta
extrair o arquivo e executar o programa LiveQuest.

### Configuração das plataformas

Na versão beta, é necessário uma série de etapas para configuração,
e a execução via linha comandos para se observar erros/links que possam
aparecer.

1. Criar um App na sua conta da Twitch e exportar o ClientID e ClientSecret.
2. Criar um projeto no Youtube e configurar o OAuth Consent Screen, e então
   obter o ClientID e ClientSecret.
3. Exportar variáveis com os valores obtidos nos passo 1 e 2.
   YOUTUBE_CLIENT_ID, YOUTUBE_CLIENT_SECRET, TWITCH_CLIENT_ID e TWITCH_CLIENT_SECRET
4. Na primeira execução, será necessário autorizar via OAuth 2 e o programa
   exibirá um link para este procedimento no terminal de comandos.

Para configurar qual o canal/chat da live do Youtube, é necessário passar
os parâmetros apropriados na linha de comandos:

```
  --twitch-channel *meu_canal*
  --youtube-stream *id_do_chat*
```

## Funcionalidades 

- Suporte às plataformas do Youtube e Twitch para *live stream*.
- Histórico do chat para integração ao OBS Studio.
- Avatar de Gopher para os expectadores, com customização de cores.
- Progresso em XP por interações e resolução de desafios de programação.

## Planejamento

Estamos planejando em adicionar estes recursos:

### Poderes e Evoluções dos Gophers 

- Hadouken
- Shoryuken
- Came-hame-ha
- Spear do Scorpion

### Aparêcias (roupinhas)

- Inspirado no Mario
- Inspirado no Sonic
- Ninja inspirado no Naruto

## Referências

1. API do Youtube para transmissão ao vivo: https://developers.google.com/youtube/v3/live/docs
2. API da Twitch para desenvolver um chat bot: https://dev.twitch.tv/docs/irc/
3. Documentação do Ebitengine: https://ebitengine.org/

## Ferramentas utilizadas para o desenvolvimento

* Vim (raiz!) para a edição do código, com auxílio do plugin `vim-go` e algumas
  outras configurações.
* Ebitengine como motor de jogo simples em Go.
* Biblioteca do Youtube para Go para obter as interações do chat.
