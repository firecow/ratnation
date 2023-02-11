import assert from "assert";
import {IncomingMessage, ServerResponse} from "http";
import rawBody from "raw-body";
import {Logger} from "../logger.mjs";
import {State} from "../state-handler.mjs";
import {Provisioner} from "./provisioner.mjs";

export default async function putKing (logger: Logger, req: IncomingMessage, res: ServerResponse, state: State, provisioner: Provisioner) {
    const body = await rawBody(req);
    const data = JSON.parse(`${body}`);
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
        return res.end("host field cannot be null or undefined\n");
    }

    if (data["ready_service_ids"] == null) {
        res.statusCode = 400;
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        return res.end("readyServices field cannot be null or undefined\n");
    }

    for (const serviceId of data["ready_service_ids"]) {
        const service = state.services.find(s => s["service_id"] === serviceId);
        assert(service != null, "service is undefined or null");
        if (!service.king_ready) {
            service.king_ready = true;
            state.revision++;
            provisioner.provision();
        }
    }

    for (const rathole of data["ratholes"]) {
        const king = state.kings.find(k => k.ports === rathole.ports && k.host === data.host);
        if (king) {
            if (king.shutting_down !== data["shutting_down"]) {
                king.shutting_down = data["shutting_down"];
                state.revision++;
                provisioner.provision();
            }
            king.beat = Date.now();
            continue;
        }
        state.kings.push({
            bind_port: rathole["bind_port"],
            ports: rathole["ports"],
            host: data["host"],
            location: data["location"],
            beat: Date.now(),
            shutting_down: false,
        });
        state.revision++;
        provisioner.provision();
    }

    res.setHeader("Content-Type", "text/plain; charset=utf-8");
    res.end("ok");
}
