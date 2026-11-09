import http from 'k6/http';
import { check, sleep } from 'k6';

export const options = {
    vus: 10,        // 10 usuarios virtuales concurrentes
    duration: '5s', // Prueba de estrés de 5 segundos
};

export default function () {
    const params = {
        headers: {
            'X-User-ID': '1', // Usamos el ID del usuario real para que evalúe amistades y gustos
            'Content-Type': 'application/json',
        },
    };

    // 1. Prueba de carga: Algoritmo Reciente (Cronológico)
    let resRecent = http.get('http://localhost:8084/v1/feed?algorithm=recent', params);
    check(resRecent, {
        'Feed cronológico cargado (status 200)': (r) => r.status === 200,
        'Respuesta incluye data': (r) => JSON.parse(r.body).data !== undefined,
    });

    sleep(0.5);

    // 2. Prueba de carga: Algoritmo Relevante (Popularidad + Preferencias)
    let resRelevant = http.get('http://localhost:8084/v1/feed?algorithm=relevant', params);
    check(resRelevant, {
        'Feed relevante cargado (status 200)': (r) => r.status === 200,
        'Respuesta incluye data': (r) => JSON.parse(r.body).data !== undefined,
    });

    sleep(0.5);
}