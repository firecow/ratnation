import assert from "assert";
import crypto from "crypto";
import {RouteCtx} from "./council-server.js";
import {streamToString} from "../utils.js";

export default async function ({req, res, state, provisioner}: RouteCtx) {
    const body = await streamToString(req);
    assert(body.length > 0, "no json data received");
    const data = JSON.parse(`${body}`);
    assert(data["ratholes"] != null, "ratholes field cannot be null or undefined");
    assert(data["preferred_location"] != null, "preferred_location field cannot be null or undefined");
    assert(data["shutting_down"] != null, "shutting_down field cannot be null or undefined");
    assert(data["ready_service_ids"] != null, "ready_service_ids field cannot be null or undefined");

    for (const serviceId of data["ready_service_ids"]) {
        const service = state.services.find(s => s["service_id"] === serviceId);
        assert(service != null, "service is undefined or null");
        if (!service.ling_ready) {
            service.ling_ready = true;
            state.revision++;
            provisioner.provision();
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
            provisioner.provision();
        }

        const service = state.services.find(s => s["name"] === rathole["name"] && s["ling_id"] === data["ling_id"]);
        if (service) {
            res.setHeader("Content-Type", "text/plain; charset=utf-8");
            res.end("ok");
            return;
        }

        const token = `${crypto.randomBytes(20).toString("hex")}`;
        state.services.push({
            service_id: crypto.randomUUID(),
            name: rathole["name"],
            token: token,
            preferred_location: data["preferred_location"],
            ling_id: data["ling_id"],
            ling_ready: false,
            remote_port: null,
            host: null,
            bind_port: null,
            king_ready: false,
        });
        state.revision++;
        provisioner.provision();
        res.setHeader("Content-Type", "text/plain; charset=utf-8");
        res.end("ok");
    }
}
