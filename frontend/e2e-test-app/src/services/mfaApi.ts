/**
 * MFA API Service for E2E Testing App
 */

import { apiClient, handleApiError } from '../../../shared/utils/api-client';
import { API_ENDPOINTS } from '../../../shared/constants/api';
import type {
  MFAEnrollResponse,
  MFAVerifyRequest,
  MFAChallengeRequest,
  MFAVerifyChallengeRequest,
} from '../../../shared/types/api';

export const mfaApi = {
  enroll: async (): Promise<MFAEnrollResponse> => {
    const response = await apiClient.post<MFAEnrollResponse>(
      API_ENDPOINTS.MFA.ENROLL
    );
    return response.data;
  },

  verify: async (data: MFAVerifyRequest): Promise<void> => {
    await apiClient.post(API_ENDPOINTS.MFA.VERIFY, data);
  },

  challenge: async (data: MFAChallengeRequest): Promise<{ challenge_id: string }> => {
    const response = await apiClient.post<{ challenge_id: string }>(
      API_ENDPOINTS.MFA.CHALLENGE,
      data
    );
    return response.data;
  },

  verifyChallenge: async (data: MFAVerifyChallengeRequest): Promise<{ access_token: string }> => {
    const response = await apiClient.post<{ access_token: string }>(
      API_ENDPOINTS.MFA.VERIFY_CHALLENGE,
      data
    );
    return response.data;
  },
};

export { handleApiError };

