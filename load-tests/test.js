import { check, sleep } from "k6";
import http from "k6/http";

const MAX_VUS = 1000;

export const options = {
  stages: [
    { duration: "20s", target: MAX_VUS },
    { duration: "10s", target: MAX_VUS },
    { duration: "20s", target: 0 },
  ],
  thresholds: {
    http_req_duration: ["p(95)<500"],
    http_req_failed: ["rate<0.1"],
  },
};

const BASE_URL = "http://nginx:4000";
const USERS = [...Array(MAX_VUS).keys()].map((i) => ({ userID: i + 1 }));

export default function () {
  const user = USERS[Math.floor(Math.random() * USERS.length)];

  // Ping endpoint
  let res = http.get(`${BASE_URL}/ping`);
  check(res, { "ping status was 200": (r) => r.status === 200 });

  // Login
  res = http.post(
    `${BASE_URL}/login`,
    JSON.stringify({ user_id: user.userID }),
    { headers: { "Content-Type": "application/json" } }
  );
  check(res, { "login status was 200": (r) => r.status === 200 });
  const tokens = res.json();

  // Authenticate
  res = http.get(`${BASE_URL}/authenticate`, {
    headers: { Authorization: `Bearer ${tokens.access}` },
  });
  check(res, { "authenticate status was 200": (r) => r.status === 200 });

  // Refresh token
  res = http.post(
    `${BASE_URL}/refresh`,
    JSON.stringify({ refresh: tokens.refresh }),
    { headers: { "Content-Type": "application/json" } }
  );
  check(res, { "refresh status was 200": (r) => r.status === 200 });
  const newTokens = res.json();

  // Logout
  res = http.post(`${BASE_URL}/logout`, null, {
    headers: { Authorization: `Bearer ${tokens.access}` },
  });
  check(res, { "logout status was 200": (r) => r.status === 200 });

  sleep(Math.random() * 5);
}
