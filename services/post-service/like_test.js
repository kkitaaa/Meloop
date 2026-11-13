import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 1,        // 1 único usuario virtual
    iterations: 1, // 1 sola iteración para probar el flujo secuencial
};

export default function () {
    const url = 'http://localhost:8084/v1/posts/post-1793509289910/like';
    const params = {
        headers: {
            'X-User-ID': '1', // Usamos el ID del usuario real de tu base de datos
            'Content-Type': 'application/json',
        },
    };

    // 1. Escenario: Dar Like (POST)
    let resLike = http.post(url, null, params);
    check(resLike, {
        'Like registrado exitosamente (status 200)': (r) => r.status === 200,
    });

    sleep(1);

    // 2. Escenario: Prevenir Like duplicado (POST)
    let resDuplicado = http.post(url, null, params);
    check(resDuplicado, {
        'Like duplicado rechazado (status 409)': (r) => r.status === 409,
    });

    sleep(1);

    // 3. Escenario: Quitar Like (DELETE)
    let resUnlike = http.del(url, null, params);
    check(resUnlike, {
        'Like eliminado exitosamente (status 200)': (r) => r.status === 200,
    });
}