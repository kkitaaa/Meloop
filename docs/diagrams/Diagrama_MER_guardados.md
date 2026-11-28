# Extensión del MER: publicaciones guardadas

Este fragmento complementa el MER entregado (`Diagrama_MER.jpg`) con la
relación de guardados. La clave primaria compuesta evita duplicar el guardado
de una misma publicación por usuario.

```mermaid
erDiagram
    USUARIO {
        varchar id_usuario PK
    }
    PUBLICACION {
        varchar id_publicacion PK
    }
    saved_posts {
        varchar user_id PK, FK
        varchar post_id PK, FK
        timestamptz saved_at
    }

    USUARIO ||--o{ saved_posts : guarda
    PUBLICACION ||--o{ saved_posts : es_guardada
```
