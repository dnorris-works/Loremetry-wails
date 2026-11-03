let currentUser = null;
const listeners = new Set();
function notify() {
    for (const fn of listeners)
        fn();
}
export function subscribeAuth(fn) {
    listeners.add(fn);
    return () => {
        listeners.delete(fn);
    };
}
export function isAuthenticated() {
    return currentUser !== null;
}
export function getCurrentUser() {
    return currentUser;
}
export function setCurrentUser(user) {
    currentUser = user;
    notify();
}
export function signOut() {
    currentUser = null;
    notify();
}
