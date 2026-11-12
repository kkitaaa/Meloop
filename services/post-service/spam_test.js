import http from 'k6/http';
import { check } from 'k6';

export const options = {
    vus: 1, 
    iterations: 55, 
};

export default function () {
    const url = 'http://localhost:8084/api/v1/posts/post-123/comments';
    
    // Ahora el JSON coincide perfectamente con el struct CreateCommentRequest
    const payload = JSON.stringify({ 
        author_id: 'user_spammer', 
        content: 'Spam test!'
    });
    
    const params = {
        headers: {
            'Content-Type': 'application/json',
            'X-User-ID': 'user_spammer', 
        },
    };

    let res = http.post(url, payload, params);

    check(res, {
        'permitido (200 o 201)': (r) => r.status === 200 || r.status === 201,
        'bloqueado por spam (429)': (r) => r.status === 429,
    });
}