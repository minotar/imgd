import http from 'k6/http';
import { randomItem } from 'https://jslib.k6.io/k6-utils/1.1.0/index.js';
import { sleep } from 'k6';

const users = [
    'clone1018',
    'd9135e082f2244c89cb0bee234155292',
    'LukeHandle',
    '5c115ca73efd41178213a0aff8ef11e0',
    'connor4312',
    'lukegb',
    '2f3665cc5e29439bbd14cb6d3a6313a7',
    'citricsquid',
    '48a0a7e4d5594873a617dc189f76a8a1',
    'KamalN',
    'Notch',
    'Banxsi',
    'externo6',
    'drupal',
    'ez',
    'd9135e082f2244c89cb0bee234100000',
    'd9135e082f2244c89cb0bee234100001',
    'd9135e082f2244c89cb0bee234100002',
    'd9135e082f2244c89cb0bee234100003',
    'd9135e082f2244c89cb0bee234100004',
    'd9135e082f2244c89cb0bee234100005'
];


export default function () {
    const randomUser = randomItem(users);

    sleep(1);
    http.get(`http://skind:4643/skin/${randomUser}`);
}
