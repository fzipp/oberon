// Copyright 2021 Frederik Zipp. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

document.addEventListener("DOMContentLoaded", function () {
    "use strict";

    const canvases = document.getElementsByTagName("canvas");
    for (let i = 0; i < canvases.length; i++) {
        const canvas = canvases[i];
        const config = configFrom(canvas.dataset);
        if (config.drawUrl) {
            webSocketCanvas(canvas, config);
            if (config.contextMenuDisabled) {
                disableContextMenu(canvas);
            }
        }
    }

    function pollClipboardChange(onChange) {
        let clipboardText = '';
        return setInterval(function() {
            navigator.clipboard.readText().then(function(clipText) {
                if (clipText !== clipboardText) {
                    clipboardText = clipText;
                    onChange({data: clipText});
                }
            });
        }, 1000);
    }

    function configFrom(dataset) {
        return {
            drawUrl: absoluteWebSocketUrl(dataset["websocketDrawUrl"]),
            eventMask: parseInt(dataset["websocketEventMask"], 10) || 0,
            reconnectInterval: parseInt(dataset["websocketReconnectInterval"], 10) || 0,
            contextMenuDisabled: (dataset["disableContextMenu"] === "true")
        };
    }

    function absoluteWebSocketUrl(url) {
        if (!url) {
            return null;
        }
        if (url.indexOf("ws://") === 0 || url.indexOf("wss://") === 0) {
            return url;
        }
        const wsUrl = new URL(url, window.location.href);
        wsUrl.protocol = wsUrl.protocol.replace("http", "ws");
        return wsUrl.href;
    }

    function webSocketCanvas(canvas, config) {
        const ctx = canvas.getContext("2d");
        const webSocket = new WebSocket(config.drawUrl);
        let handlers = {};
        webSocket.binaryType = "arraybuffer";
        webSocket.addEventListener("open", function () {
            handlers = addEventListeners(canvas, config.eventMask, webSocket);
        });
        webSocket.addEventListener("error", function () {
            webSocket.close();
        });
        webSocket.addEventListener("close", function () {
            removeEventListeners(canvas, handlers);
            if (!config.reconnectInterval) {
                return;
            }
            setTimeout(function () {
                webSocketCanvas(canvas, config);
            }, config.reconnectInterval);
        });
        webSocket.addEventListener("message", function (event) {
            draw(ctx, new DataView(event.data));
        });
    }

    function addEventListeners(canvas, eventMask, webSocket) {
        const handlers = {};

        if (eventMask & 1) {
            handlers["mousemove"] = sendMouseEvent(1);
        }
        if (eventMask & 2) {
            handlers["mousedown"] = sendMouseEvent(2);
        }
        if (eventMask & 4) {
            handlers["mouseup"] = sendMouseEvent(3);
        }
        if (eventMask & 8) {
            handlers["keydown"] = sendKeyEvent(4);
        }
        if (eventMask & 16) {
            handlers["keyup"] = sendKeyEvent(5);
        }
        if (eventMask & 32) {
            handlers["click"] = sendMouseEvent(6);
        }
        if (eventMask & 64) {
            handlers["dblclick"] = sendMouseEvent(7);
        }
        if (eventMask & 128) {
            handlers["auxclick"] = sendMouseEvent(8);
        }
        if (eventMask & 256) {
            handlers["wheel"] = sendWheelEvent(9);
        }
        if (eventMask & 512) {
            handlers["touchstart"] = sendTouchEvent(10);
        }
        if (eventMask & 1024) {
            handlers["touchmove"] = sendTouchEvent(11);
        }
        if (eventMask & 2048) {
            handlers["touchend"] = sendTouchEvent(12);
        }
        if (eventMask & 4096) {
            handlers["touchcancel"] = sendTouchEvent(13);
        }
        if (eventMask & 8192) {
            pollClipboardChange(sendClipboardEvent(14));
        }

        Object.keys(handlers).forEach(function (type) {
            const target = targetFor(type, canvas);
            target.addEventListener(type, handlers[type], {passive: false});
        });

        const rect = canvas.getBoundingClientRect();

        const mouseMoveThreshold = 25;
        let lastMouseMoveTime = -1;

        function sendMouseEvent(eventType) {
            return function (event) {
                event.preventDefault();
                if (eventType === 1) {
                    const now = new Date().getTime();
                    if ((now - lastMouseMoveTime) < mouseMoveThreshold) {
                        return;
                    }
                    lastMouseMoveTime = now;
                }
                const eventMessage = new ArrayBuffer(11);
                const dataView = new DataView(eventMessage);
                setMouseEvent(dataView, eventType, event);
                webSocket.send(eventMessage);
            };
        }

        function sendWheelEvent(eventType) {
            return function (event) {
                event.preventDefault();
                const eventMessage = new ArrayBuffer(36);
                const dataView = new DataView(eventMessage);
                setMouseEvent(dataView, eventType, event);
                dataView.setFloat64(11, event.deltaX);
                dataView.setFloat64(19, event.deltaY);
                dataView.setFloat64(27, event.deltaZ);
                dataView.setUint8(35, event.deltaMode);
                webSocket.send(eventMessage);
            };
        }

        function setMouseEvent(dataView, eventType, event) {
            dataView.setUint8(0, eventType);
            dataView.setUint8(1, event.buttons);
            dataView.setUint32(2, ((event.clientX - rect.left) / canvas.offsetWidth) * canvas.width);
            dataView.setUint32(6, ((event.clientY - rect.top) / canvas.offsetHeight) * canvas.height);
            dataView.setUint8(10, encodeModifierKeys(event));
        }

        function sendTouchEvent(eventType) {
            return function (event) {
                event.preventDefault();
                const touchBytes = 12;
                const eventMessage = new ArrayBuffer(1 +
                    1 + (event.touches.length * touchBytes) +
                    1 + (event.changedTouches.length * touchBytes) +
                    1 + (event.targetTouches.length * touchBytes) +
                    1);
                const dataView = new DataView(eventMessage);
                let offset = 0;
                dataView.setUint8(offset, eventType);
                offset++;
                offset = setTouches(dataView, offset, event.touches);
                offset = setTouches(dataView, offset, event.changedTouches);
                offset = setTouches(dataView, offset, event.targetTouches);
                dataView.setUint8(offset, encodeModifierKeys(event));
                webSocket.send(eventMessage);
            };
        }

        function setTouches(dataView, offset, touches) {
            const len = touches.length;
            dataView.setUint8(offset, len);
            offset++;
            for (let i = 0; i < len; i++) {
                const touch = touches[i];
                dataView.setUint32(offset, touch.identifier);
                offset += 4;
                dataView.setUint32(offset, ((touch.clientX - rect.left) / canvas.offsetWidth) * canvas.width);
                offset += 4;
                dataView.setUint32(offset, ((touch.clientY - rect.top) / canvas.offsetHeight) * canvas.height);
                offset += 4;
            }
            return offset;
        }

        function sendKeyEvent(eventType) {
            return function (event) {
                event.preventDefault();
                const keyBytes = new TextEncoder().encode(event.key);
                const eventMessage = new ArrayBuffer(6 + keyBytes.byteLength);
                const data = new DataView(eventMessage);
                data.setUint8(0, eventType);
                data.setUint8(1, encodeModifierKeys(event));
                data.setUint32(2, keyBytes.byteLength);
                for (let i = 0; i < keyBytes.length; i++) {
                    data.setUint8(6 + i, keyBytes[i]);
                }
                webSocket.send(eventMessage);
            };
        }

        function sendClipboardEvent(eventType) {
            return function (event) {
                const dataBytes = new TextEncoder().encode(event.data);
                const eventMessage = new ArrayBuffer(5 + dataBytes.byteLength);
                const data = new DataView(eventMessage);
                data.setUint8(0, eventType);
                data.setUint32(1, dataBytes.byteLength);
                for (let i = 0; i < dataBytes.length; i++) {
                    data.setUint8(5 + i, dataBytes[i]);
                }
                webSocket.send(eventMessage);
            };
        }

        return handlers;
    }

    function removeEventListeners(canvas, handlers) {
        Object.keys(handlers).forEach(function (type) {
            const target = targetFor(type, canvas);
            target.removeEventListener(type, handlers[type]);
        });
    }

    function targetFor(eventType, canvas) {
        if ((eventType.indexOf("key") !== 0) && (eventType.indexOf("composition") !== 0)) {
            return canvas;
        }
        return document;
    }

    function disableContextMenu(canvas) {
        canvas.addEventListener("contextmenu", function (e) {
            e.preventDefault();
        }, false);
    }

    function encodeModifierKeys(event) {
        let modifiers = 0;
        if (event.altKey) {
            modifiers |= 1;
        }
        if (event.shiftKey) {
            modifiers |= 2;
        }
        if (event.ctrlKey) {
            modifiers |= 4;
        }
        if (event.metaKey) {
            modifiers |= 8;
        }
        return modifiers;
    }

    function draw(ctx, data) {
        switch (data.getUint8(0)) {
            case 1:
                const x = data.getUint32(1);
                const y = data.getUint32(5);
                const width = data.getUint32(9);
                const height = data.getUint32(13);
                const len = width * height * 4;
                const bufferOffset = data.byteOffset + 17;
                const buffer = data.buffer.slice(bufferOffset, bufferOffset + len);
                const array = new Uint8ClampedArray(buffer);
                const imageData = new ImageData(array, width, height);
                ctx.putImageData(imageData, x, y);
                return 17 + len;
            case 2:
                const text = getString(data, 1);
                navigator.clipboard.writeText(text.value);
                return 1 + text.byteLen;
        }
        return 1;
    }
});

function getString(data, offset) {
    const stringLen = data.getUint32(offset);
    const stringBegin = data.byteOffset + offset + 4;
    const stringEnd = stringBegin + stringLen;
    return {
        value: new TextDecoder().decode(data.buffer.slice(stringBegin, stringEnd)),
        byteLen: 4 + stringLen
    };
}
