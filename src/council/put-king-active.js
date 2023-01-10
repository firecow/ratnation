import rawBody from "raw-body";

export default async function putKingActive(req, res, state) {
    const body = await rawBody(req);
    const data = JSON.parse(body);
    const names = data["names"];


    const services = state.services.filter(s => names.includes(s["name"]));
    const inactives = services.filter(s => s["king_active"] === false);
    if (inactives.length > 0) state.revision++;
    services.forEach(s => s["king_active"] = true);

    res.setHeader("Content-Type", "text/plain; charset=utf-8");
    res.end("ok");
}
