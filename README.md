# jogo-da-live

![Gopher](cmd/jogo-da-live/assets/img/gopher_standing.gif)

Este é o nosso primeiro projeto na Codigolândia: um jogo desenvolvido em **Go**
e que pode ser jogado durante as lives!

Neste jogo cada participante terá seu avatar exibido como um **overlay** sobre
a live, e as interações no chat poderão ser utilizadas para movimentar o seu
personagem, atacar outros jogadores e muito mais!

Poderemos também adicionar as mensagens do chat nesta implementação e também
recursos exclusivos para membros/apoiadores do canal.

## Ferramentas utilizadas

* Vim (raiz!) para a edição do código, com auxílio do plugin `vim-go` e algumas
  outras configurações.
* Ebitengine como motor de jogo simples em Go.
* Biblioteca do Youtube para Go para obter as interações do chat.

## Como funciona?

O programa em `cmd/jogo-da-live` irá desenhar uma janela com um fundo verde,
que é adicionada ao OBS e possui o filtro de *chroma key* aplicado.

Este programa se conecta aos chats da transmissão ao vivo e utiliza esta
informação para desenhar os personagens
na tela.

## TODO

- [x] Ler os expectadores do Youtube
- [x] Ler os expectadores do Twitch
- [x] Enviar o PONG ao receber um PING da Twitch
- [x] Preservar o histórico do chat
- [x] Não duplicar os expectadores
- [x] Criar um pixel art do Gopher
- [ ] Movimentar o expectador na tela
- [ ] Adicionar os comandos de chat

### Poderes e Powerups

- Hadouken
- Shoryuken
- Came-hame-ha
- Spear do Scorpion

### Skins

- Inspirado no Mario
- Inspirado no Sonic
- Ninja inspirado no Naruto

## Referências

1. API do Youtube para transmissão ao vivo: https://developers.google.com/youtube/v3/live/docs
2. API da Twitch para desenvolver um chat bot: https://dev.twitch.tv/docs/irc/
3. Documentação do Ebitengine: https://ebitengine.org/
