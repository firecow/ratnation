import assert from "assert";
import {RouteCtx} from "./council-server.mjs";
import {streamToString} from "../utils.mjs";

export default async function ({req, res, state, provisioner}: RouteCtx) {
    const body = await streamToString(req);
    assert(body.length > 0, "no json data received");
    const data = JSON.parse(`${body}`);
    assert(data["ratholes"] != null, "ratholes field cannot be null or undefined");
    assert(data["location"] != null, "location field cannot be null or undefined");
    assert(data["host"] != null, "host field cannot be null or undefined");
    assert(data["ready_service_ids"] != null, "ready_service_ids field cannot be null or undefined");

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
