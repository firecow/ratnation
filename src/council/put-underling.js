import rawBody from "raw-body";
import crypto from "crypto";

export default async function putUnderling(req, res, state) {
    const body = await rawBody(req);
    const data = JSON.parse(body);
    if (data["ratholes"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("ratholes field cannot be null or undefined\n");
    }
    if (data["prefered_location"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("prefered_location field cannot be null or undefined\n");
    }

    for (const rathole of data["ratholes"]) {

        const service = state.services.find(s => s["name"] === rathole["name"]);
        if (service) {
            service["underling_ping"] = Date.now();
            res.setHeader("Content-Type", "text/plain; charset=utf-8");
            return res.end("ok");
        }

        const token = `${crypto.randomBytes(20).toString("hex")}`;
        state.services.push({
            name: rathole["name"],
            token: token,
            prefered_location: data["prefered_location"],
            underling_ping: Date.now(),
            location: null,
            host: null,
            remote_port: null,
            bind_port: null,
            king_active: false,
        });
        state.revision++;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        res.end("ok");
    }
}
