from __future__ import annotations

import google.generativeai as genai

from application.interfaces import Llm


class GeminiLlm(Llm):
    def __init__(self, *, api_key: str, model: str):
        genai.configure(api_key=api_key)
        self._model = genai.GenerativeModel(model)

    def generate_json_answer(self, prompt: str) -> str:
        response = self._model.generate_content(
            prompt,
            generation_config={"temperature": 0.2},
        )
        text = getattr(response, "text", None)
        if text:
            return str(text)
        if response.candidates:
            parts = response.candidates[0].content.parts
            return "".join(getattr(p, "text", "") for p in parts)
        return ""
