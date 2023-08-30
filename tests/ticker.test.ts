import {afterEach, beforeEach, expect, jest, test} from "@jest/globals";
import {Ticker} from "../src/ticker.js";
import waitForExpect from "wait-for-expect";

let ticker: Ticker;
let tickMock: () => Promise<void>;
beforeEach(() => {
    tickMock = jest.fn<() => Promise<void>>();
});

afterEach(() => {
    ticker?.stop();
});

test("ticker.tick()", async () => {
    ticker = new Ticker({
        interval: 500,
        tick: tickMock,
    });
    await ticker.tick();

    expect(tickMock).toHaveBeenCalledTimes(1);
});

test("ticker.start()", async () => {
    ticker = new Ticker({
        interval: 500,
        tick: tickMock,
    });
    ticker.start();

    // eslint-disable-next-line @typescript-eslint/ban-ts-comment
    // @ts-ignore
    await waitForExpect(() => expect(tickMock).toHaveBeenCalled());
});

test("ticker.stop()", () => {
    ticker = new Ticker({
        interval: 500,
        tick: tickMock,
    });
    ticker.stop();
});
