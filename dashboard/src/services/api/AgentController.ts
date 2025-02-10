import { request } from '@umijs/max';

export async function proposals(
  body: AGENTAPI.ProposalsReq,
  options?: { [key: string]: any },
) {
  return request<AGENTAPI.ProposalsRes>('/api/proposals', {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
    },
    data: body,
    ...(options || {}),
  });
}


export async function manifesto(
  options?: { [key: string]: any },
) {
  return request<AGENTAPI.ManifestoRes>('/api/manifesto', {
    method: 'GET',
    ...(options || {}),
  });
}
