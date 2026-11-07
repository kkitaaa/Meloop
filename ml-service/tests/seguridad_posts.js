import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 1,
    duration: '2s',
};

// NOTA: Ajusta el puerto si tu post-service corre en uno distinto (ej. 8080)
const BASE_URL = 'http://localhost:8080/api/v1/posts';

export default function () {
    const fakeToken = 'Bearer token-falso-para-pruebas';
    const fakePostId = 'uuid-inventado-1234';

    const params = {
        headers: {
            'Content-Type': 'application/json',
            'Authorization': fakeToken, // Tu middleware debería interceptar esto
        },
    };

    // 1. Intentar editar un post inexistente
    const payloadEdit = JSON.stringify({ content: "Texto editado" });
    let resEdit = http.put(`${BASE_URL}/${fakePostId}`, payloadEdit, params);
    
    check(resEdit, {
        'Edición bloqueada correctamente (404 o 403)': (r) => r.status === 404 || r.status === 403 || r.status === 401,
    });

    // 2. Intentar eliminar un post inexistente
    let resDel = http.del(`${BASE_URL}/${fakePostId}`, null, params);
    
    check(resDel, {
        'Eliminación bloqueada correctamente (404 o 403)': (r) => r.status === 404 || r.status === 403 || r.status === 401,
    });

    sleep(1);
}