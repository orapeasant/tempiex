"""Payload conversion utilities - JSON encoding/decoding of workflow/activity arguments and results."""
import json


def encode_payloads(values: list):
    """Convert Python values to Payloads proto message."""
    from tempiex.api.common.v1 import message_pb2
    payloads = message_pb2.Payloads()
    for v in values:
        p = message_pb2.Payload()
        p.encoding = "json/plain"
        p.data = json.dumps(v).encode("utf-8")
        payloads.payloads.append(p)
    return payloads


def decode_payloads(payloads_proto) -> list:
    """Convert Payloads proto to Python values."""
    if payloads_proto is None:
        return []
    results = []
    for p in payloads_proto.payloads:
        results.append(json.loads(p.data.decode("utf-8")))
    return results


def decode_single(payloads_proto):
    """Decode a single value from Payloads."""
    values = decode_payloads(payloads_proto)
    if not values:
        return None
    return values[0]
