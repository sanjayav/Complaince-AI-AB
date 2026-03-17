 import hashlib
 import math
 import os
 from typing import Iterable, List

 import numpy as np

 try:
     from langchain_community.embeddings import OpenAIEmbeddings
 except Exception:  # pragma: no cover
     OpenAIEmbeddings = None  # type: ignore

 DEFAULT_DIM = 1536


 def _deterministic_embedding(text: str, dim: int = DEFAULT_DIM) -> List[float]:
     # Fallback embedding: hash-based deterministic vector for dev without API keys
     h = hashlib.sha256(text.encode("utf-8")).digest()
     rng_seed = int.from_bytes(h[:8], "big")
     rng = np.random.default_rng(rng_seed)
     vec = rng.standard_normal(dim)
     vec = vec / np.linalg.norm(vec)
     return vec.astype(float).tolist()


 class EmbeddingModel:
     def __init__(self, dim: int = DEFAULT_DIM):
         self.dim = dim
         self.provider = os.getenv("EMBEDDINGS_PROVIDER", "openai")
         self.openai_key = os.getenv("OPENAI_API_KEY")
         self._client = None
         if self.provider == "openai" and self.openai_key and OpenAIEmbeddings is not None:
             self._client = OpenAIEmbeddings(model="text-embedding-3-small")

     def embed(self, text: str) -> List[float]:
         if self._client is not None:
             return self._client.embed_query(text)
         return _deterministic_embedding(text, self.dim)

     def embed_many(self, texts: Iterable[str]) -> List[List[float]]:
         return [self.embed(t) for t in texts]

