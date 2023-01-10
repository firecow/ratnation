export default async function getState(req, res, state) {
    res.setHeader("Content-Type", "application/json; charset=utf-8");
    res.end(JSON.stringify(state));
}
