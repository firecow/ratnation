import {beforeEach, describe, expect, test} from "@jest/globals";
import request from "supertest";
import createServer from "../src/council-server.js";
import {CouncilProvisioner} from "../src/council-provisioner.js";
import {Logger} from "../src/logger.js";
import {State} from "../src/state-handler.js";

let provisioner: CouncilProvisioner, state: State, logger: Logger;
beforeEach(() => {
    state = {kings: [], services: [], revision: 0, lings: []};
    logger = new Logger();
    provisioner = new CouncilProvisioner({logger});
});

test("GET /state", async () => {
    const server = createServer({provisioner, state}).httpServer;
    const res = await request(server).get("/state");
    expect(res.statusCode).toEqual(200);
    expect(res.body).toEqual({kings: [], services: [], revision: 0, lings: []});
});

test("GET /not-found", async () => {
    const server = createServer({provisioner, state}).httpServer;
    const res = await request(server).get("/not-found");
    expect(res.statusCode).toEqual(404);
    expect(res.text).toEqual("Page could not be found");
});

describe("PUT /king", () => {
    test("empty body", async () => {
        const server = createServer({provisioner, state}).httpServer;
        const res = await request(server).put("/king");
        expect(res.statusCode).toEqual(400);
        expect(res.text).toEqual("no json data received");
    });

    test("success", async () => {
        const server = createServer({provisioner, state}).httpServer;
        const res = await request(server).put("/king").send({
            ratholes: [],
            ready_service_ids: [],
            location: "mylocation",
            host: "example.com",
        });
        expect(res.text).toEqual("ok");
        expect(res.statusCode).toEqual(200);
    });
});

