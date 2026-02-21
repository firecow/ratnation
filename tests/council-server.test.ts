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

    test("stores noise_public_key on king creation", async () => {
        const server = createServer({provisioner, state}).httpServer;
        await request(server).put("/king").send({
            ratholes: [{bind_port: 2333, ports: "5000-5001"}],
            ready_service_ids: [],
            location: "mylocation",
            host: "example.com",
            noise_public_key: "abc123pubkey",
        });
        expect(state.kings).toHaveLength(1);
        expect(state.kings[0].noise_public_key).toEqual("abc123pubkey");
    });

    test("stores null noise_public_key when not provided", async () => {
        const server = createServer({provisioner, state}).httpServer;
        await request(server).put("/king").send({
            ratholes: [{bind_port: 2333, ports: "5000-5001"}],
            ready_service_ids: [],
            location: "mylocation",
            host: "example.com",
        });
        expect(state.kings).toHaveLength(1);
        expect(state.kings[0].noise_public_key).toBeNull();
    });

    test("updates noise_public_key on existing king", async () => {
        state.kings.push({
            bind_port: 2333,
            ports: "5000-5001",
            host: "example.com",
            location: "mylocation",
            beat: 0,
            shutting_down: false,
            noise_public_key: null,
        });
        const initialRevision = state.revision;
        const server = createServer({provisioner, state}).httpServer;
        await request(server).put("/king").send({
            ratholes: [{bind_port: 2333, ports: "5000-5001"}],
            ready_service_ids: [],
            location: "mylocation",
            host: "example.com",
            noise_public_key: "newkey123",
        });
        expect(state.kings[0].noise_public_key).toEqual("newkey123");
        expect(state.revision).toBeGreaterThan(initialRevision);
    });
});

