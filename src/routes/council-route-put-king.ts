import assert from "assert";
import {RouteCtx} from "../council-server.js";
import {streamToString} from "../utils.js";

export default async function ({req, res, state, provisioner, socketIo}: RouteCtx) {
    const body = await streamToString(req);
    assert(body.length > 0, "no json data received");
    const data = JSON.parse(`${body}`);
    assert(data["ratholes"] != null, "ratholes field cannot be null or undefined");
    assert(data["location"] != null, "location field cannot be null or undefined");
    assert(data["host"] != null, "host field cannot be null or undefined");
    assert(data["ready_service_ids"] != null, "ready_service_ids field cannot be null or undefined");

    for (const serviceId of data["ready_service_ids"]) {
        const service = state.services.find(s => s["service_id"] === serviceId);
        assert(service != null, `${serviceId} cannot be found in state.services`);
        if (!service.king_ready) {
            service.king_ready = true;
            state.revision++;
            provisioner.provision(state);
            socketIo.sockets.emit("state-changed");
        }
    }

    for (const rathole of data["ratholes"]) {
        const king = state.kings.find(k => k.ports === rathole.ports && k.host === data.host);
        if (king) {
            if (king.shutting_down !== data["shutting_down"]) {
                king.shutting_down = data["shutting_down"];
                state.revision++;
                provisioner.provision(state);
                socketIo.sockets.emit("state-changed");
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
        provisioner.provision(state);
        socketIo.sockets.emit("state-changed");
    }

    res.setHeader("Content-Type", "text/plain; charset=utf-8");
    res.end("ok");
}
