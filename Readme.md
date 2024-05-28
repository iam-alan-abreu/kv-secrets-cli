
# Aplicação de Extração de Segredos do Azure Key Vault

Este repositório contém uma aplicação Go que autentica no Azure Key Vault, extrai segredos e os salva em um arquivo `.env` para ser utilizado por outras aplicações.

## Pré-requisitos

Antes de executar a aplicação, você precisa ter:

1. **Conta no Azure**: Certifique-se de ter uma conta no Azure.
2. **Azure Key Vault**: Um Key Vault configurado no Azure.
3. **Go**: Instale a linguagem Go na versão 1.19 ou superior.
4. **Docker**: Para rodar a aplicação em containers.
5. **Azure CLI**: Instale a Azure CLI para configurar as permissões(opcional).

## Configuração do Ambiente

### 1. Configurar Permissões no Portal do Azure

#### Passo 1: Configurar Permissões no Portal do Azure

Acesse o Portal do Azure:

### 2. Configurar Permissões no Key Vault

Garanta que o `usuario` que irá utilizar a aplicação tenha permissão para acessar os segredos no Key Vault. As permissões necessárias são `get` e `list` para segredos.

```sh
az keyvault set-policy --name <NOME_DO_VAULT> --object-id <USER_OBJECT_ID> --secret-permissions get list
```

Para pesquisar qual o object id do usuario basta usar o comando abaixo:

```sh
az ad user show --id <USER_EMAIL_OR_UPN> --query objectId --output tsv
```

### 3. Armazenar Segredos no Key Vault

Os segredos no Key Vault precisam estar armazenados de acordo com a seguinte convenção:

- **Sem prefixo de aplicação**: Se o segredo não tiver um prefixo de aplicação, armazene-o no formato `<NOME-DA-CHAVE>`.
- **Com prefixo de aplicação**: Se o segredo tiver um prefixo de aplicação, armazene-o no formato `<NOME-APP>-<NOME-DA-CHAVE>`.

Por exemplo, se o nome da aplicação for `myapp` e o nome do segredo for `dbpassword`, o segredo deve ser armazenado como `MYAPP-DBPASSWORD`.


## Instalação e Execução

### 1. Clonar o Repositório

Clone este repositório para sua máquina local:

```sh
git clone <URL_DO_REPOSITORIO>
cd <NOME_DO_REPOSITORIO>
```

### 2. Construir a Aplicação

Construa a aplicação Go:

```sh
go build -o keyvault-secrets
```

### 3. Executar a Aplicação

Execute a aplicação passando o nome do Key Vault e, opcionalmente, o nome do aplicativo para filtrar os segredos:

```sh
./keyvault-secrets <NOME_DO_VAULT> <NOME_DO_APP>
```

- `<NOME_DO_VAULT>`: O nome do Azure Key Vault onde estão armazenados os segredos.
- `<NOME_DO_APP>` (opcional): O prefixo dos segredos para a aplicação específica.

Isso criará um arquivo `.env` no diretório `output` com os segredos extraídos.

## Docker

Para rodar a aplicação usando Docker, utilize os arquivos `Dockerfile` e `docker-compose.yml` fornecidos.

### Construir a Imagem Docker

Construa a imagem Docker:

```sh
docker-compose build
```

### Executar os Containers

Execute os containers:

```sh
docker-compose up
```

## Contribuição

Sinta-se à vontade para contribuir para este projeto enviando pull requests ou abrindo issues. Certifique-se de seguir o [código de conduta](CODE_OF_CONDUCT.md) e as diretrizes de contribuição.


## Referencias

https://ravichaganti.com/blog/azure-sdk-for-go-authentication-methods-for-local-dev-environment/