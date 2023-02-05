import rawBody from "raw-body";
import crypto from "crypto";

export default async function putling(req, res, state, provisioner) {
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
    if (data["ready_service_ids"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("ready_service_ids field cannot be null or undefined\n");
    }
    if (data["shutting_down"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("shutting_down field cannot be null or undefined\n");
    }

    for (const serviceId of data["ready_service_ids"]) {
        const service = state.services.find(s => s["service_id"] === serviceId);
        if (!service.ling_ready) {
            service.ling_ready = true;
            state.revision++;
            await provisioner.provision();
        }
    }

    for (const rathole of data["ratholes"]) {

        let ling = state.lings.find(u => u["ling_id"] === data["ling_id"]);
        if (!ling) {
            ling = {ling_id: data["ling_id"], beat: Date.now(), shutting_down: data["shutting_down"]};
            state.lings.push(ling);
        }
        ling.beat = Date.now();
        if (ling.shutting_down !== data["shutting_down"]) {
            ling.shutting_down = data["shutting_down"];
            state.revision++;
            await provisioner.provision();
        }

        const service = state.services.find(s => s["name"] === rathole["name"] && s["ling_id"] === data["ling_id"]);
        if (service) {
            res.setHeader("Content-Type", "text/plain; charset=utf-8");
            return res.end("ok");
        }

        const token = `${crypto.randomBytes(20).toString("hex")}`;
        state.services.push({
            service_id: crypto.randomUUID(),
            name: rathole["name"],
            token: token,
            prefered_location: data["prefered_location"],
            ling_id: data["ling_id"],
            ling_ready: false,
            remote_port: null,
            host: null,
            bind_port: null,
            king_ready: false,
        });
        state.revision++;
        await provisioner.provision();
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        res.end("ok");
    }
}
