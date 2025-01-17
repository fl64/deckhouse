---
title: "Модуль user-authz: настройки"
---

{% include module-bundle.liquid %}

> **Внимание!** Мы категорически не рекомендуем создавать Pod'ы и ReplicaSet'ы – эти объекты являются второстепенными и должны создаваться из других контроллеров. Доступ к созданию и изменению Pod'ов и ReplicaSet'ов полностью отсутствует.
>
> **Внимание!** Режим multi-tenancy (авторизация по namespace) в данный момент реализован по временной схеме и **не гарантирует безопасность**! Если webhook, который реализовывает систему авторизации по какой-то причине упадёт, авторизация по namespace (опции `allowAccessToSystemNamespaces` и `limitNamespaces` в CR) перестанет работать и пользователи получат доступы во все namespace. После восстановления доступности webhook'а все вернется на свои места.

## Параметры

<!-- SCHEMA -->

Вся настройка прав доступа происходит с помощью [Custom Resources](cr.html).
