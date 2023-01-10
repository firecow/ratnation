import rawBody from "raw-body";

export default async function putKing(req, res, state) {
    const body = await rawBody(req);
    const data = JSON.parse(body);
    if (data["ratholes"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("ratholes field cannot be null or undefined\n");
    }
    if (data["location"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("location field cannot be null or undefined\n");
    }

    if (data["host"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("location field cannot be null or undefined\n");
    }

    for (const rathole of data["ratholes"]) {
        const king = state.kings.find(k => k["ports"] === rathole["ports"] && k["host"] === data["host"]);
        if (king) {
            king.ping = Date.now();
            continue;
        }
        state.kings.push({
            bind_port: rathole["bind_port"],
            ports: rathole["ports"],
            used: [],
            host: data["host"],
            location: data["location"],
            ping: Date.now(),
            shutting_down: false,
        });
        state.revision++;
    }

    res.setHeader("Content-Type", "text/plain; charset=utf-8");
    res.end("ok");
}
