"""
Command autopilot exposes an API to decide what to do in the game.
"""

from pathlib import Path
from typing import Tuple

import torch, uvicorn
from fastapi import FastAPI, Body, HTTPException
from pydantic import BaseModel, Field

from stig.internal.env import env
from stig.internal.dataset.image import process_from_bytes, to_tensor
from stig.internal.game.action import THROTTLE_VALUES_MAP, STEERING_VALUES_MAP

class ActResp(BaseModel):
    throttle: str
    steering: str

def create_app(
    model_name: str,
    models_dir: str,
    frame_size: Tuple[int, int],
    device: str,
):
    """
    Autopilot creates an API server to interact with the AI model.
    """
    # Create the model path.
    model_path = Path(models_dir) / f"{model_name}.pt"
    if not model_path.exists():
        raise FileNotFoundError(f"Model {model_path} not found")

    # Load the model.
    model = torch.jit.load(str(model_path)).to(device)
    model.eval()

    # Create FastAPI app.
    app = FastAPI(title="Stig Autopilot")

    @app.post("/act")
    async def act(
        img_bytes: bytes = Body(..., media_type="image/jpeg")
    ) -> ActResp:
        # Process the image.
        try:
            img = process_from_bytes(img_bytes, frame_size, device)
        except Exception as e:
            raise HTTPException(status_code=400, detail=str(e))

        # Predict the action.
        try:
            tensor = to_tensor(img, device)
            throttle_logits, steering_logits = model(tensor)
            throttle_id = int(throttle_logits.argmax())
            steering_id = int(steering_logits.argmax())
        except Exception as e:
            raise HTTPException(status_code=500, detail=str(e))

        # Return the action.
        return ActResp(
            throttle=THROTTLE_VALUES_MAP[throttle_id],
            steering=STEERING_VALUES_MAP[steering_id],
        )

    return app

if __name__ == "__main__":
    # Read environment variables.
    model_name = env.lookup("MODEL_NAME")

    models_dir = env.lookup("MODELS_DIR")

    frame_height = env.lookupInt("FRAME_HEIGHT")
    frame_width = env.lookupInt("FRAME_WIDTH")
    frame_size = (frame_height, frame_width)

    api_host = env.lookup("API_HOST")
    api_port = env.lookupInt("API_PORT")

    device = "cuda" if torch.cuda.is_available() else "cpu"

    # Create the server.
    app = create_app(
        model_name,
        models_dir,
        frame_size,
        device,
    )

    # Run the server.
    uvicorn.run(app, host=api_host, port=api_port)
