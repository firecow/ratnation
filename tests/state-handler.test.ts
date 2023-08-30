import {beforeEach, expect, jest, test} from "@jest/globals";
import {State, StateHandler} from "../src/state-handler.js";
import {Logger} from "../src/logger.js";

let stateHandler: StateHandler;
let logger: Logger;
let stateChangedMock;

beforeEach(() => {
    stateChangedMock = jest.fn<(state: State) => Promise<void> | void>();
    logger = {info: jest.fn(), error: jest.fn()};
    stateHandler = new StateHandler({logger, councilHost: "ratcouncil.example.io", stateChanged: stateChangedMock});
});


test("Handles bad status code", async () => {
    await stateHandler.fetchState();

    expect(logger.error).toHaveBeenCalledWith("Failed to fetch state from council", expect.objectContaining({"error.message": "Invalid URL"}));
});
