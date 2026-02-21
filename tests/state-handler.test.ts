import {afterEach, beforeEach, expect, jest, test} from "@jest/globals";
import {State, StateHandler} from "../src/state-handler.js";
import {Logger} from "../src/logger.js";

let stateHandler: StateHandler;
let logger: Logger;
let stateChangedMock: ReturnType<typeof jest.fn>;
let errorMock: ReturnType<typeof jest.fn>;

beforeEach(() => {
    stateChangedMock = jest.fn<(state: State) => Promise<void> | void>();
    errorMock = jest.fn();
    logger = {info: jest.fn(), error: errorMock} as unknown as Logger;
    stateHandler = new StateHandler({logger, councilHost: "ratcouncil.example.io", stateChanged: stateChangedMock});
});

afterEach(() => {
    stateHandler.stop();
});

test("Handles bad status code", async () => {
    await stateHandler.fetchState();

    expect(errorMock).toHaveBeenCalledWith("Failed to fetch state from council", expect.objectContaining({"error.message": "Invalid URL"}));
});
