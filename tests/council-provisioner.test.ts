import {beforeEach, expect, jest, test} from "@jest/globals";
import {CouncilProvisioner} from "../src/council-provisioner.js";
import {Logger} from "../src/logger.js";
import {State} from "../src/state-handler.js";

let provisioner: CouncilProvisioner;
let logger: Logger;

beforeEach(() => {
    logger = {info: jest.fn(), error: jest.fn()};
});


test("Find available port on king", () => {
    const state: State = {
        services: [
            {
                service_id: "some_service_id",
                ling_id: "someid",
                name: "alpha",
                token: "some_token",
                ling_ready: false,
                king_ready: false,
                preferred_location: "myhouse",
                host: null,
                bind_port: null,
                remote_port: null,
            },
        ],
        kings: [{host: "kinghost.com", ports: "5000-5000", location: "myhouse", bind_port: 2343, beat: 0, shutting_down: false}],
        lings: [{ling_id: "some_ling_id", beat: 0, shutting_down: false}],
        revision: 0,
    };
    provisioner = new CouncilProvisioner({logger});

    const kingPorts = provisioner.availableKingPorts(state);
    expect(kingPorts).toEqual([
        {
            "king": {
                "beat": 0,
                "bind_port": 2343,
                "host": "kinghost.com",
                "location": "myhouse",
                "ports": "5000-5000",
                "shutting_down": false,
            },
            "ports": [5000],
        },
    ]);
});
