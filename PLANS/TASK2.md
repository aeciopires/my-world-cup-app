# My World Cup App

Crie uma aplicação web usando Golang para exibir a tabela da copa do mundo de 2026 e informações complementares da FIFA.

# Additional Features

Implemente as funcionalidades adicionais:

- Path de healthcheck
- Path de metrics 
- Crie o helm chart. Todos os helm values devem estar documentados em ingles en-US.
- Deve ser utilizado o helm-docs (https://github.com/norwoodj/helm-docs) para atualizar a doc do helm chart usando o template https://github.com/aeciopires/helm-watchdog-pod-delete/blob/main/charts/helm-watchdog-pod-delete/README.md.gotmpl
- Uma pagina para mostrar estatisticas por jogador, selecao, gols marcados
- Arquivo de CONTRIBUTING.md semelhante a https://raw.githubusercontent.com/aeciopires/learning-istio/refs/heads/main/CONTRIBUTING.md
- No README.md liste todos os requisitos de software para desenvolver e executar o software
- No arquivo Makefile deve ter uma opcao para verificar se todos as dependecias de software estao instaladas. Quando nao, deve listar quais precisam ser instaladas
- No README.md deve ter uma tabela com todas as variaveis de ambiente (se houverem)
- No Makefile deve ter opcao para instalar a aplicacao usando o helm, atualizar doc do helm chart, fazer lint do helm chart
- Atualize todos os testes e garanta que tudo funcione conforme especificado

Requisitos não funcionais:

- Teste o deploy da aplicacao via helm no cluster kind-king-multinodes
- Todo o código e documentação deve estar em ingles en-US
- Utilize as versões mais novas das tecnologias
- Utilize as melhores praticas de codificação, system design, clean code
- Testes unitarios
- Dockerfile e Docker Compose para cada aplicação
- Arquivo Makefile para facilitar a execução da aplicação e execução de testes
- Atualize o arquivo README.md com instruções de uso, da arquitetura, componentes de software utilizados, workflow em mermaid, estrutura de diretórios
- Um arquivo de CHANGELOG.md
- Um arquivo CLAUDE.md com informações relevantes ao projeto

