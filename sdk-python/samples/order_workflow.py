"""
Order Workflow — sample demonstrating all Tempiex SDK annotations.

Annotations shown:
  @workflow.defn          — marks a class as a workflow
  @workflow.run           — marks the main entry-point coroutine
  @workflow.signal        — handles an inbound signal (async, mutates state)
  @workflow.query         — answers a synchronous state query (no side effects)
  @workflow.update        — synchronous RPC that mutates state and returns a value
  @activity.defn          — marks a function as an activity

Run this sample:
    # Terminal 1 — start the Tempiex server (if not already running)
    /home/ubuntu/app/tempiex/tempiex/bin/tempiex-server

    # Terminal 2 — run the worker + client
    cd /home/ubuntu/app/tempiex/sdk-python
    .venv/bin/python samples/order_workflow.py
"""
import asyncio
import uuid
from datetime import timedelta

from tempiex.client import Client
from tempiex.worker import Worker
from tempiex import workflow, activity


# ---------------------------------------------------------------------------
# Activities
# ---------------------------------------------------------------------------

@activity.defn
async def validate_order(order_id: str, amount: float) -> dict:
    """Validate an order — returns the validated order dict."""
    print(f"[activity] validate_order: order={order_id} amount={amount}")
    if amount <= 0:
        raise ValueError(f"Invalid amount: {amount}")
    return {"order_id": order_id, "amount": amount, "validated": True}


@activity.defn
async def charge_payment(order_id: str, amount: float) -> str:
    """Charge payment — returns a transaction ID."""
    print(f"[activity] charge_payment: order={order_id} amount={amount}")
    txn_id = f"txn-{uuid.uuid4().hex[:8]}"
    return txn_id


@activity.defn(name="send-confirmation-email")
async def send_confirmation(order_id: str, txn_id: str, email: str) -> bool:
    """Send a confirmation email. Custom activity name via @activity.defn(name=...)."""
    print(f"[activity] send_confirmation: order={order_id} txn={txn_id} to={email}")
    return True


# ---------------------------------------------------------------------------
# Workflow
# ---------------------------------------------------------------------------

@workflow.defn(name="OrderWorkflow")
class OrderWorkflow:
    """
    Order processing workflow with signals, queries, and updates.

    State machine:
        pending → approved/rejected (via @signal)
        running activities to charge + confirm
        supports cancellation via @signal
        exposes status and item count via @query
        supports adding items via @update
    """

    def __init__(self):
        self._status: str = "pending"
        self._approved: bool = False
        self._rejected: bool = False
        self._cancelled: bool = False
        self._rejection_reason: str = ""
        self._items: list[str] = []
        self._txn_id: str | None = None

    # --- Signal handlers ---------------------------------------------------

    @workflow.signal
    async def approve(self, approver: str = ""):
        """Signal: approve this order."""
        print(f"[signal] approve received from={approver!r}")
        self._approved = True
        self._status = "approved"

    @workflow.signal
    async def reject(self, reason: str = ""):
        """Signal: reject this order."""
        print(f"[signal] reject received reason={reason!r}")
        self._rejected = True
        self._rejection_reason = reason
        self._status = "rejected"

    @workflow.signal
    async def cancel(self):
        """Signal: cancel this order regardless of approval state."""
        print("[signal] cancel received")
        self._cancelled = True
        self._status = "cancelled"

    # --- Query handlers ----------------------------------------------------

    @workflow.query
    def get_status(self) -> str:
        """Query: return current order status."""
        return self._status

    @workflow.query
    def get_items(self) -> list[str]:
        """Query: return current item list."""
        return list(self._items)

    @workflow.query(name="item-count")
    def item_count(self) -> int:
        """Query: return item count. Custom query name via @workflow.query(name=...)."""
        return len(self._items)

    # --- Update handlers ---------------------------------------------------

    @workflow.update
    async def add_item(self, item: str) -> int:
        """Update: add an item and return the new total count."""
        print(f"[update] add_item: {item!r}")
        self._items.append(item)
        return len(self._items)

    @workflow.update
    async def remove_item(self, item: str) -> bool:
        """Update: remove an item if present, return whether it was found."""
        if item in self._items:
            self._items.remove(item)
            return True
        return False

    # --- Main entry point --------------------------------------------------

    @workflow.run
    async def run(self, order_id: str, amount: float, customer_email: str) -> dict:
        """Main workflow logic."""
        self._status = "awaiting_approval"
        print(f"[workflow] OrderWorkflow started: order={order_id} amount={amount}")

        # Wait for approval, rejection, or cancellation
        await workflow.wait_condition(
            lambda: self._approved or self._rejected or self._cancelled
        )

        if self._cancelled:
            return {"order_id": order_id, "status": "cancelled"}

        if self._rejected:
            return {
                "order_id": order_id,
                "status": "rejected",
                "reason": self._rejection_reason,
            }

        # Approved — run activities
        self._status = "processing"

        order_info = await workflow.execute_activity(
            validate_order,
            order_id,
            amount,
            start_to_close_timeout=timedelta(seconds=10),
        )

        txn_id = await workflow.execute_activity(
            charge_payment,
            order_id,
            order_info["amount"],
            start_to_close_timeout=timedelta(seconds=10),
        )
        self._txn_id = txn_id

        await workflow.execute_activity(
            send_confirmation,
            order_id,
            txn_id,
            customer_email,
            start_to_close_timeout=timedelta(seconds=10),
        )

        self._status = "completed"
        return {
            "order_id": order_id,
            "status": "completed",
            "transaction_id": txn_id,
            "items": self._items,
        }


# ---------------------------------------------------------------------------
# Runner
# ---------------------------------------------------------------------------

async def main():
    client = await Client.connect("localhost:8133", namespace="sample-orders")

    worker = Worker(
        client,
        task_queue="orders",
        workflows=[OrderWorkflow],
        activities=[validate_order, charge_payment, send_confirmation],
    )
    worker_task = asyncio.create_task(worker.run())

    order_id = f"order-{uuid.uuid4().hex[:8]}"

    try:
        print(f"\n=== Starting OrderWorkflow: {order_id} ===\n")

        # Start workflow (non-blocking)
        handle = await client.start_workflow(
            OrderWorkflow,
            order_id,
            99.99,
            "customer@example.com",
            id=order_id,
            task_queue="orders",
        )

        await asyncio.sleep(0.5)

        # --- Demonstrate @workflow.signal ---
        print("\n--- Sending 'approve' signal ---")
        await handle.signal("approve", "alice")

        # Give worker time to process
        await asyncio.sleep(1.0)

        # Wait for the result
        result = await handle.result()
        print(f"\n=== Workflow result: {result} ===\n")

    finally:
        worker.shutdown()
        await asyncio.sleep(0.3)
        worker_task.cancel()
        try:
            await worker_task
        except asyncio.CancelledError:
            pass
        await client.close()


if __name__ == "__main__":
    asyncio.run(main())
