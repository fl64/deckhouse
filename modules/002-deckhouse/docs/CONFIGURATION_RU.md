---
title: "Модуль deckhouse: настройки"
---

{% include module-bundle.liquid %}

## Параметры

<!-- SCHEMA -->

**Внимание!** В случае, если в `nodeSelector` указан несуществующий лейбл или указаны неверные `tolerations`, Deckhouse перестанет работать. Для восстановления работоспособности необходимо изменить значения на правильные в `configmap/deckhouse` и в `deployment/deckhouse`.
